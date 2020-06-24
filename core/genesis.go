// Copyright 2014 The go-auc Authors
// This file is part of the go-auc library.
//
// The go-auc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-auc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-auc library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ether-ark/etherark/crypto"
	"github.com/ether-ark/etherark/p2p/enode"
	"math/big"
	"strings"

	"github.com/ether-ark/etherark/common"
	"github.com/ether-ark/etherark/common/hexutil"
	"github.com/ether-ark/etherark/common/math"
	"github.com/ether-ark/etherark/core/rawdb"
	"github.com/ether-ark/etherark/core/state"
	"github.com/ether-ark/etherark/core/types"
	"github.com/ether-ark/etherark/ethdb"
	"github.com/ether-ark/etherark/log"
	"github.com/ether-ark/etherark/params"
	"github.com/ether-ark/etherark/rlp"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *params.ChainConfig `json:"config"`
	Nonce      uint64              `json:"nonce"`
	Timestamp  uint64              `json:"timestamp"`
	ExtraData  []byte              `json:"extraData"`
	GasLimit   uint64              `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int            `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash         `json:"mixHash"`
	Coinbase   common.Address      `json:"coinbase"`
	StateRoot  common.Hash         `json:"stateRoot"`
	Alloc      GenesisAlloc        `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db ethdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.CircumChainConfig, common.Hash{}, errGenesisNoConfig
	}
	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, params.GenesisBlockNumber)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.Commit(db)
		return genesis.Config, block.Hash(), err
	}
	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}
	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)

		return newcfg, stored, nil
	}
	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	if genesis == nil && stored != params.MainnetGenesisHash {
		return storedcfg, stored, nil
	}
	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {

		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {

		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	default:
		return params.CircumChainConfig
	}
}

// ToBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToBlock(db ethdb.Database) *types.Block {
	if db == nil {
		db = ethdb.NewMemDatabase()
	}

	statedb, _ := state.New(g.StateRoot, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance, big.NewInt(1))
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	root := statedb.IntermediateRoot(false)

	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Nonce:      types.EncodeNonce(g.Nonce),
		Time:       g.Timestamp,
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		Difficulty: g.Difficulty,
		MixDigest:  g.Mixhash,
		Coinbase:   g.Coinbase,
		Root:       root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	if g.Difficulty == nil {
		head.Difficulty = params.GenesisDifficulty
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, false)
	block := types.NewBlock(head, nil, nil, nil)

	return block
}

// Commit writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) Commit(db ethdb.Database) (*types.Block, error) {
	block := g.ToBlock(db)
	if block.NumberU64() != params.GenesisBlockNumber {
		return nil, fmt.Errorf("can't commit genesis block with number != %d", params.GenesisBlockNumber)
	}
	rawdb.WriteTd(db, block.Hash(), block.NumberU64(), g.Difficulty)
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())

	config := g.Config
	if config == nil {
		config = params.AllEthashProtocolChanges
	}
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// MustCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustCommit(db ethdb.Database) *types.Block {
	block, err := g.Commit(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db ethdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: GenesisAlloc{addr: {Balance: balance}}}
	return g.MustCommit(db)
}

// DefaultGenesisBlock returns the Ethereum main net genesis block.
func DefaultGenesisBlock() *Genesis {
	alloc := decodePrealloc(mainnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.MainnetMasternodes)
	alloc[common.HexToAddress("0x7c0886920566489632f5D4d28f7ea9dCE8B3c95c")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x2d386F38b717120AE94bde8990f1c9Aa44932C09")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x698235D113a5ABBc34b3c4af6e23e403B4b1c1c7")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0xF4355704B2B2fc80c419558D99D06cB6458E91e9")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x8B19fFEaA8128b6d1B807BD575fF68933AeEbbFa")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0xF6221cb5bA55De8a68aBa4514a0C98A46F7aDBA3")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x8e2d8a8901f68BF8f01f000309995f770439Fb6a")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x313E56e28D0616179BC2E01f37EfD902f7Bb5209")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0x319AB84A80C5DA42159b52261470258b740e12f3")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	alloc[common.HexToAddress("0xd9e3BaFBAAc75c773F876d3E135b91f0C2B7C0Ec")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(5000e+7), big.NewInt(1e+15)),
	}
	config := params.CircumChainConfig
	var witnesses []string
	for _, n := range params.MainnetMasternodes {
		node := enode.MustParseV4(n)
		pubkey := node.Pubkey()
		addr := crypto.PubkeyToAddress(*pubkey)
		if _, ok := alloc[addr]; !ok {
			alloc[addr] = GenesisAccount{
				Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e+16)),
			}
		}
		xBytes := pubkey.X.Bytes()
		var x [32]byte
		copy(x[32-len(xBytes):], xBytes[:])
		id1 := common.BytesToHash(x[:])
		id := fmt.Sprintf("%x", id1[:8])
		witnesses = append(witnesses, id)
	}
	config.Circum.Witnesses = witnesses
	return &Genesis{
		Config:     config,
		Nonce:      1,
		Timestamp:  1583712800,
		GasLimit:   10000000,
		Difficulty: big.NewInt(1),
		Alloc:      alloc,
		Number:     params.GenesisBlockNumber,
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	alloc := decodePrealloc(testnetAllocData)
	alloc[common.BytesToAddress(params.MasterndeContractAddress.Bytes())] = masternodeContractAccount(params.TestnetMasternodes)
	alloc[common.HexToAddress("0x4b961Cc393e08DF94F70Cad88142B9962186FfD1")] = GenesisAccount{
		Balance: new(big.Int).Mul(big.NewInt(1e+11), big.NewInt(1e+15)),
	}
	config := params.TestnetChainConfig
	var witnesses []string
	for _, n := range params.TestnetMasternodes {
		node := enode.MustParseV4(n)
		pubkey := node.Pubkey()
		//addr := crypto.PubkeyToAddress(*pubkey)
		//if _, ok := alloc[addr]; !ok {
		//	alloc[addr] = GenesisAccount{
		//		Balance: new(big.Int).Mul(big.NewInt(1e+16), big.NewInt(1e+15)),
		//	}
		//}
		xBytes := pubkey.X.Bytes()
		var x [32]byte
		copy(x[32-len(xBytes):], xBytes[:])
		id1 := common.BytesToHash(x[:])
		id := fmt.Sprintf("%x", id1[:8])
		witnesses = append(witnesses, id)
	}
	config.Circum.Witnesses = witnesses
	return &Genesis{
		Config:     config,
		Nonce:      66,
		Timestamp:  1531551970,
		ExtraData:  hexutil.MustDecode("0x3535353535353535353535353535353535353535353535353535353535353535"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(1048576),
		Alloc:      alloc,
	}
}

// DefaultRinkebyGenesisBlock returns the Rinkeby network genesis block.
func DefaultRinkebyGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.RinkebyChainConfig,
		Timestamp:  1492009146,
		ExtraData:  hexutil.MustDecode("0x52657370656374206d7920617574686f7269746168207e452e436172746d616e42eb768f2244c8811c63729a21a3569731535f067ffc57839b00206d1ad20c69a1981b489f772031b279182d99e65703f0076e4812653aab85fca0f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc:      decodePrealloc(rinkebyAllocData),
	}
}

// DefaultGoerliGenesisBlock returns the Görli network genesis block.
func DefaultGoerliGenesisBlock() *Genesis {
	return &Genesis{
		Config:     params.GoerliChainConfig,
		Timestamp:  1548854791,
		ExtraData:  hexutil.MustDecode("0x22466c6578692069732061207468696e6722202d204166726900000000000000e0a2bd4258d2768837baa26a28fe71dc079f84c70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		GasLimit:   10485760,
		Difficulty: big.NewInt(1),
		// Alloc:      decodePrealloc(goerliAllocData),
	}
}

// DeveloperGenesisBlock returns the 'geth --dev' genesis block. Note, this must
// be seeded with the
func DeveloperGenesisBlock(period uint64, faucet common.Address) *Genesis {
	// Override the default period to the user requested one
	config := *params.AllCliqueProtocolChanges
	config.Clique.Period = period

	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, 65)...),
		GasLimit:   6283185,
		Difficulty: big.NewInt(1),
		Alloc: map[common.Address]GenesisAccount{
			common.BytesToAddress([]byte{1}): {Balance: big.NewInt(1)}, // ECRecover
			common.BytesToAddress([]byte{2}): {Balance: big.NewInt(1)}, // SHA256
			common.BytesToAddress([]byte{3}): {Balance: big.NewInt(1)}, // RIPEMD
			common.BytesToAddress([]byte{4}): {Balance: big.NewInt(1)}, // Identity
			common.BytesToAddress([]byte{5}): {Balance: big.NewInt(1)}, // ModExp
			common.BytesToAddress([]byte{6}): {Balance: big.NewInt(1)}, // ECAdd
			common.BytesToAddress([]byte{7}): {Balance: big.NewInt(1)}, // ECScalarMul
			common.BytesToAddress([]byte{8}): {Balance: big.NewInt(1)}, // ECPairing
			faucet:                           {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
		},
	}
}

func decodePrealloc(data string) GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = GenesisAccount{Balance: account.Balance}
	}
	return ga
}

func masternodeContractAccount(masternodes []string) GenesisAccount {
	data := make(map[common.Hash]common.Hash)

	data[common.HexToHash("0x8f83dea0634358fc7f838269b688e7442ff3056f3b24689e2f9ec5dc24058984")] = common.HexToHash("0x000000000000000000000000000000001b655173cceef89559364f156ff4043e")
	data[common.HexToHash("0x5d9a27bfd5dc79d60516adaf9b5601dcb8ecfc92317656b268b71412a530eb2e")] = common.HexToHash("0x0000000000000000000000000000000079fabe943d8df0251b655173cceef895")
	data[common.HexToHash("0xc7e097067f482fae2b5fa795ab3c7b473803ab44b417e1e8f592b4cf18b5f87a")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa6576aa5945acee866420644a993fc1de3252e13c7cb25bbcc9e905e88f88c00")] = common.HexToHash("0x0605906618803bca951c490eddde3bc4b5770026f74f8d8021cb407a7c8cdf72")
	data[common.HexToHash("0x5d9a27bfd5dc79d60516adaf9b5601dcb8ecfc92317656b268b71412a530eb2d")] = common.HexToHash("0x5605b9d443454f65524dafa6b0025cbb2cf582f0dc1a86877753c04eff0ada00")
	data[common.HexToHash("0x6ab26da15846deed161dc7b7322a2306e79804161790e4f8c3700ab98538a539")] = common.HexToHash("0x0000000000000000000000000000000059364f156ff4043e2e62177f24597ecc")
	data[common.HexToHash("0xadb2661d4b20820392bd90ddb7c0213acaf2bb4cfc9ebc641f68859a2d3356a2")] = common.HexToHash("0x0000000000000000000000000000000000000000000000009f01ed31d22f0327")
	data[common.HexToHash("0xa6576aa5945acee866420644a993fc1de3252e13c7cb25bbcc9e905e88f88c01")] = common.HexToHash("0x000000000000000000000000000000002e62177f24597ecc0000000000000000")
	data[common.HexToHash("0x6ab26da15846deed161dc7b7322a2306e79804161790e4f8c3700ab98538a53a")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x5d9a27bfd5dc79d60516adaf9b5601dcb8ecfc92317656b268b71412a530eb2c")] = common.HexToHash("0x534b6ebf40c104b8a9d3bd2b422e8a43315840256f61dbfb6105826286a0adc4")
	data[common.HexToHash("0xdefabb881945c0df93534730ce6c15efcc6256fd01373746688b84ea73e3f29a")] = common.HexToHash("0x0000000000000000000000000000000034d6a3fafeaff37b90a6be32e6840c90")
	data[common.HexToHash("0x5d9a27bfd5dc79d60516adaf9b5601dcb8ecfc92317656b268b71412a530eb2f")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x7ed7339dae2783d26228a3e8f98eeb7d1dfcb32d7afd7560a12d5d1c3c775035")] = common.HexToHash("0x90a6be32e6840c909591f8f2bb000c58251c383c2f2335937cad07610b54806f")
	data[common.HexToHash("0xbf9d2bb87e62e535ea6ef5ac0a334e29d385528e7e05f1878903c10ffbc7d30b")] = common.HexToHash("0x0000000000000000000000000000000073dd97fa443c86f282a45bc11685b08a")
	data[common.HexToHash("0xf770100495b2a545e20ee1f9a5be262b5a65491436bbfe155ae8605f066f64a5")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xe06c3fbd3dbb4752bc961c27186a104c28c8e73f97ff6a0a04020801c9543dbd")] = common.HexToHash("0xc3c0900c7a59fb2f428106f6498833a7bf80d8b63ec42033928bb0028c050550")
	data[common.HexToHash("0xc26a13dcdb1f16b39f1401be74aaf4d48cb92f7b240150b2a9550e1b93efc2f1")] = common.HexToHash("0x2e62177f24597ecc86b3b9606b2c36075a40b3f0ea3e2bb74d281721f9bf897e")
	data[common.HexToHash("0xdefabb881945c0df93534730ce6c15efcc6256fd01373746688b84ea73e3f299")] = common.HexToHash("0x900ecde3308a36a234c5b62186392a3af69c84e2f6911ff29122fd6b356df9ab")
	data[common.HexToHash("0xc26a13dcdb1f16b39f1401be74aaf4d48cb92f7b240150b2a9550e1b93efc2f3")] = common.HexToHash("0x000000000000000000000000000000005bcd5d081899a5aa459d477d048bcb41")
	data[common.HexToHash("0x6ab26da15846deed161dc7b7322a2306e79804161790e4f8c3700ab98538a538")] = common.HexToHash("0x47fcbe33b12e9bd71ee1bb647e88bf4c8a9437ff8a97099baa023bcb917f3586")
	data[common.HexToHash("0xabcf212daf70cd5ced9a5fab66287bbcbb33f4bd560d5170e5524a2b021a863e")] = common.HexToHash("0xb0b84a57c8a832c2184fac51d58588c3a86002c9e81cfb543e99789c8c4c87d7")
	data[common.HexToHash("0xc3db16296a126ae9001348cdfd0132769dd804474ebd360dbcc20bbc95cc0e6c")] = common.HexToHash("0x00000000000000000000000000000000eeb86e957f52c5dc9f01ed31d22f0327")
	data[common.HexToHash("0xf73a5cd2144ee06854ae219eda6b6f7bfce76139a745bbfe7f5df14f32465e29")] = common.HexToHash("0x000000000000000000000000000000000000000000000000f05f45d104e0f4f3")
	data[common.HexToHash("0x236479d00156ba43a63c1de7decb35686d77d7e876cb85ee0bd05fed5a344fa5")] = common.HexToHash("0xdb06d6834f932753c61f14f6fb177d3804f0a5c628c98768d57c4ccb8c5cc859")
	data[common.HexToHash("0xe06c3fbd3dbb4752bc961c27186a104c28c8e73f97ff6a0a04020801c9543dbe")] = common.HexToHash("0x000000000000000000000000000000000000000000000000db06d6834f932753")
	data[common.HexToHash("0xc26a13dcdb1f16b39f1401be74aaf4d48cb92f7b240150b2a9550e1b93efc2f4")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa44c02e00dc03e073b4753b8a3ea47386a1f4ee1bae76613f29e90c162187a2b")] = common.HexToHash("0x34d6a3fafeaff37b00aa380f6c08316648d0e416d84d244d4993f65c6c6ff25f")
	data[common.HexToHash("0xa2d5f1809db342eae7d1355030756f064b23d2fe9e1a61582355eb60ded82922")] = common.HexToHash("0xf18f1612d5f2642b4be7e80ddaa8e92772978b1f99d4cf7297cf324b56d2d5dc")
	data[common.HexToHash("0xd6a49b9980cb65cf5ff7ed14b5de91db3485a6bc6366ee985a6578c5e2d68211")] = common.HexToHash("0x000000000000000000000000000000000000000000000000db06d6834f932753")
	data[common.HexToHash("0x6ab26da15846deed161dc7b7322a2306e79804161790e4f8c3700ab98538a537")] = common.HexToHash("0x5bcd5d081899a5aa088f78b0a056c90f107b5ec61295b40aa79383eb0f824d3c")
	data[common.HexToHash("0xc3db16296a126ae9001348cdfd0132769dd804474ebd360dbcc20bbc95cc0e6a")] = common.HexToHash("0x09a2e1e78b97f39a26bde28115c1138969e76fa865f4ceee69467c3f35203589")
	data[common.HexToHash("0x2ca1f376d3c6e52201eead87848701a521ddf61d1a66a80535c9b6b96d4c139e")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xf770100495b2a545e20ee1f9a5be262b5a65491436bbfe155ae8605f066f64a3")] = common.HexToHash("0x02d8afaf888d716766c2329d46ad4fdf698c3bb1b21098071cfac539092efe66")
	data[common.HexToHash("0xe06c3fbd3dbb4752bc961c27186a104c28c8e73f97ff6a0a04020801c9543dbc")] = common.HexToHash("0x8f9ef24296521239a38bf7352fe37ac82d5c3c9a8473d1778b0378e723b9cc62")
	data[common.HexToHash("0xdefabb881945c0df93534730ce6c15efcc6256fd01373746688b84ea73e3f298")] = common.HexToHash("0xcbf753979832898e2394b19313ec8bdf3fc594e5c9a62133a875af97f78a511d")
	data[common.HexToHash("0xa44c02e00dc03e073b4753b8a3ea47386a1f4ee1bae76613f29e90c162187a2d")] = common.HexToHash("0x000000000000000000000000000000009f01ed31d22f0327cbf753979832898e")
	data[common.HexToHash("0xa2d5f1809db342eae7d1355030756f064b23d2fe9e1a61582355eb60ded82923")] = common.HexToHash("0x0000000000000000000000000000000082a45bc11685b08a09a2e1e78b97f39a")
	data[common.HexToHash("0xf11f5f800ff1aa5e2f13ecae7d003277c06c6071793345c2c0df459956ad950d")] = common.HexToHash("0x00000000000000000000000000000000000000000000000073dd97fa443c86f2")
	data[common.HexToHash("0x543ff72e98d2b0090a476cf5b953ff7e11956f7d0465fe2e80df46a5b62714ca")] = common.HexToHash("0xb5d364080df08c1782b7ef2770d56249f460a740e2a353671e54e85e47c7f2ea")
	data[common.HexToHash("0xabcf212daf70cd5ced9a5fab66287bbcbb33f4bd560d5170e5524a2b021a863d")] = common.HexToHash("0x1b655173cceef895b57abba85df4e14a5d6627d4b30e508f587f61e16ce056b7")
	data[common.HexToHash("0x73c140488b5c857f1ff7c6873e4da9d78e6eedc05be4338c6dff457b49c0fb29")] = common.HexToHash("0x0000000000000000000000000000000000000000000000001b655173cceef895")
	data[common.HexToHash("0x7ed7339dae2783d26228a3e8f98eeb7d1dfcb32d7afd7560a12d5d1c3c775037")] = common.HexToHash("0x00000000000000000000000000000000cbf753979832898e79fabe943d8df025")
	data[common.HexToHash("0xe8d68415012443f7b78a5f83e537ac4be3283fb02d12b1bda9cdd6ee181e29b5")] = common.HexToHash("0x000000000000000000000000000000000000000000000000b5d364080df08c17")
	data[common.HexToHash("0x886f08cbfc8104dd01a1082eb98794097b2183825189e9e6fb68a89913047120")] = common.HexToHash("0x0000000000000000000000000000000090a6be32e6840c90534b6ebf40c104b8")
	data[common.HexToHash("0xc3db16296a126ae9001348cdfd0132769dd804474ebd360dbcc20bbc95cc0e6b")] = common.HexToHash("0x324921873ef646b40060e3f783cab89e8c0832f7cd87351c9130175ab7d62a12")
	data[common.HexToHash("0x543ff72e98d2b0090a476cf5b953ff7e11956f7d0465fe2e80df46a5b62714cc")] = common.HexToHash("0x00000000000000000000000000000000db06d6834f932753b7982598f229847d")
	data[common.HexToHash("0x543ff72e98d2b0090a476cf5b953ff7e11956f7d0465fe2e80df46a5b62714cd")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x8f83dea0634358fc7f838269b688e7442ff3056f3b24689e2f9ec5dc24058982")] = common.HexToHash("0x4c2ae630ee2b7324a137eda69ae667c95d3db7d7c9b252e6e59a470561453343")
	data[common.HexToHash("0xbf9d2bb87e62e535ea6ef5ac0a334e29d385528e7e05f1878903c10ffbc7d309")] = common.HexToHash("0xf05f45d104e0f4f382a43a9eb5c68c1477de4210d08df3ee07cac59fd7b40972")
	data[common.HexToHash("0x236479d00156ba43a63c1de7decb35686d77d7e876cb85ee0bd05fed5a344fa6")] = common.HexToHash("0x1187f014163472584a69eb15d8d6971e017153b03895c1c6e3901c3e19d6229f")
	data[common.HexToHash("0xc26a13dcdb1f16b39f1401be74aaf4d48cb92f7b240150b2a9550e1b93efc2f2")] = common.HexToHash("0xe103cbe9237421dcf66247c15352409a32ba0f4a0895a9018598cfe857da9e88")
	data[common.HexToHash("0x5968fb10a8706789e4f92ec0dfca0fbb41d77a75c86984bdadc8387971d3f60f")] = common.HexToHash("0x000000000000000000000000000000000000000000000000534b6ebf40c104b8")
	data[common.HexToHash("0x37ade56ddeaf07eea736ce3ce05fe8f3a94b058dea597121dca9fe26e0124771")] = common.HexToHash("0x82a45bc11685b08afc97984ea532f2be8c4c9ca9fe66702288624f1ed5b44871")
	data[common.HexToHash("0xf770100495b2a545e20ee1f9a5be262b5a65491436bbfe155ae8605f066f64a4")] = common.HexToHash("0x00000000000000000000000000000000b5d364080df08c1773dd97fa443c86f2")
	data[common.HexToHash("0xa333a773275d4c99d76900eb51dc1ee0de4eda272707089273f79d6deae4c966")] = common.HexToHash("0x00000000000000000000000000000000000000000000000059364f156ff4043e")
	data[common.HexToHash("0x8f83dea0634358fc7f838269b688e7442ff3056f3b24689e2f9ec5dc24058985")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xabcf212daf70cd5ced9a5fab66287bbcbb33f4bd560d5170e5524a2b021a8640")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x2ca1f376d3c6e52201eead87848701a521ddf61d1a66a80535c9b6b96d4c139d")] = common.HexToHash("0x00000000000000000000000000000000b7982598f229847df05f45d104e0f4f3")
	data[common.HexToHash("0x886f08cbfc8104dd01a1082eb98794097b2183825189e9e6fb68a89913047121")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa44c02e00dc03e073b4753b8a3ea47386a1f4ee1bae76613f29e90c162187a2c")] = common.HexToHash("0x5af5de7f726e39d9e5858c7f224f3a6d3a8195f31400321bff462441f4e5ec1d")
	data[common.HexToHash("0x543ff72e98d2b0090a476cf5b953ff7e11956f7d0465fe2e80df46a5b62714cb")] = common.HexToHash("0x76f94936ba99437608f25e778e3acb107de90d3e82a4df0f5e13fa4327cea791")
	data[common.HexToHash("0x7ce3dd0876f4ad511c058fe9983396ef9b0567c4f7ceaaa0afd0d34763010b7e")] = common.HexToHash("0x0000000000000000000000000000000000000000000000008f9ef24296521239")
	data[common.HexToHash("0x886f08cbfc8104dd01a1082eb98794097b2183825189e9e6fb68a8991304711e")] = common.HexToHash("0x79fabe943d8df025672fc0866ebe0a9a219cf4b60fcd230749cf3e6769d1543b")
	data[common.HexToHash("0x5dca906c0c2670fb69fcd35608569aac3b741a4316ac13b781cb9ec28cdba46d")] = common.HexToHash("0x00000000000000000000000000000000000000000000000009a2e1e78b97f39a")
	data[common.HexToHash("0xe06c3fbd3dbb4752bc961c27186a104c28c8e73f97ff6a0a04020801c9543dbf")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xebef750ced604b5daae073d925ac7bd8f8c094775e20e098bacb2a987822c6df")] = common.HexToHash("0x0000000000000000000000000000000000000000000000004c2ae630ee2b7324")
	data[common.HexToHash("0x7ed7339dae2783d26228a3e8f98eeb7d1dfcb32d7afd7560a12d5d1c3c775038")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa44c02e00dc03e073b4753b8a3ea47386a1f4ee1bae76613f29e90c162187a2e")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x1a7584bed9ce96db21dcd36f632ed6339ed773c3aa8e57671f36ceb350e142ba")] = common.HexToHash("0x000000000000000000000000000000000000000000000000b7982598f229847d")
	data[common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")] = common.HexToHash("0x0000000000000000000000000000000000000000000000008f9ef24296521239")
	data[common.HexToHash("0x886f08cbfc8104dd01a1082eb98794097b2183825189e9e6fb68a8991304711f")] = common.HexToHash("0x15c81da4a45d47885ea645061d638445c4d9f87492d43afb739157dff8f8e9ab")
	data[common.HexToHash("0x4fd1827c65605c0e2269f9a5aba5819411e795c1dc0095e672d39bdad6d7ca90")] = common.HexToHash("0x000000000000000000000000000000000000000000000000459d477d048bcb41")
	data[common.HexToHash("0x9f63d50d4b3d32a7cc4876a4cdd52d3ad03d3df0c75b5ec1e32104946f1696c6")] = common.HexToHash("0x00000000000000000000000000000000000000000000000079fabe943d8df025")
	data[common.HexToHash("0xdefabb881945c0df93534730ce6c15efcc6256fd01373746688b84ea73e3f29b")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa6576aa5945acee866420644a993fc1de3252e13c7cb25bbcc9e905e88f88bff")] = common.HexToHash("0x459d477d048bcb41ee61700baa5832435962be4012ef8f6d6b9c480da5944ca0")
	data[common.HexToHash("0x61c9a702fb9b11b60ba6f7dd1d03b67f46c2c0086bc4414eadf0e7032e74b4f3")] = common.HexToHash("0x000000000000000000000000000000000000000000000000eeb86e957f52c5dc")
	data[common.HexToHash("0xc7e097067f482fae2b5fa795ab3c7b473803ab44b417e1e8f592b4cf18b5f878")] = common.HexToHash("0xd6e63da96054cab9ca7998d71027874a777ad2e1bfc7197c53c9ffe036964b00")
	data[common.HexToHash("0x07342186b6b8e78ca7aff0bdf9efcbdb5ba906bd3ef2d77063906c9b8e37e9e2")] = common.HexToHash("0x00000000000000000000000000000000000000000000000034d6a3fafeaff37b")
	data[common.HexToHash("0x37ade56ddeaf07eea736ce3ce05fe8f3a94b058dea597121dca9fe26e0124773")] = common.HexToHash("0x00000000000000000000000000000000f05f45d104e0f4f3eeb86e957f52c5dc")
	data[common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")] = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000015")
	data[common.HexToHash("0x312c2af045da8c58d2fbfa1d52d94ee3d6dbf2185b7b957172e9d2f9a4c8e269")] = common.HexToHash("0x000000000000000000000000000000000000000000000000cbf753979832898e")
	data[common.HexToHash("0x37ade56ddeaf07eea736ce3ce05fe8f3a94b058dea597121dca9fe26e0124772")] = common.HexToHash("0x97ca92b54473c4313fbafcd80cd37865411329d15e71a5cd2d351c3a837b5746")
	data[common.HexToHash("0xb040b6e66cd81670d00989b19acb0d885ee1d63ad05e7a61cc496e57333d535b")] = common.HexToHash("0x000000000000000000000000000000004c2ae630ee2b73245bcd5d081899a5aa")
	data[common.HexToHash("0x92e58df8080e931218e92536917f9d4337333f03097dcdadda3b4d35ee0ba685")] = common.HexToHash("0x0000000000000000000000000000000000000000000000002e62177f24597ecc")
	data[common.HexToHash("0x67b730f5789f50675298520bf0197c216a134175a94ab0fd7f27bd6f66a74f0f")] = common.HexToHash("0x0000000000000000000000000000000000000000000000005bcd5d081899a5aa")
	data[common.HexToHash("0x236479d00156ba43a63c1de7decb35686d77d7e876cb85ee0bd05fed5a344fa7")] = common.HexToHash("0x000000000000000000000000000000008f9ef24296521239b5d364080df08c17")
	data[common.HexToHash("0xa6576aa5945acee866420644a993fc1de3252e13c7cb25bbcc9e905e88f88c02")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x7a61712b386a8c1c265c5550f0fbdad7af254471837013f72c072090cad7b5a2")] = common.HexToHash("0x00000000000000000000000000000000000000000000000082a45bc11685b08a")
	data[common.HexToHash("0xbf9d2bb87e62e535ea6ef5ac0a334e29d385528e7e05f1878903c10ffbc7d30a")] = common.HexToHash("0x380a02c35d1879a5791ab193d109354501b128242387b3b7c61017fc998c257a")
	data[common.HexToHash("0xc7e097067f482fae2b5fa795ab3c7b473803ab44b417e1e8f592b4cf18b5f877")] = common.HexToHash("0x9f01ed31d22f03279445053ba759f5ddda910bc0b8c4041679983805806f24fc")
	data[common.HexToHash("0x236479d00156ba43a63c1de7decb35686d77d7e876cb85ee0bd05fed5a344fa8")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xf9dd7d618b56bc4da0a962c68e3d54dc8640fac1bc4e94580e520e7657de030e")] = common.HexToHash("0x00000000000000000000000000000000000000000000000090a6be32e6840c90")
	data[common.HexToHash("0xb040b6e66cd81670d00989b19acb0d885ee1d63ad05e7a61cc496e57333d535c")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x7ed7339dae2783d26228a3e8f98eeb7d1dfcb32d7afd7560a12d5d1c3c775036")] = common.HexToHash("0x913e15e9951d76af12092721c7c2167678a26e90ac16fdf7366da1139749bca5")
	data[common.HexToHash("0xa2d5f1809db342eae7d1355030756f064b23d2fe9e1a61582355eb60ded82921")] = common.HexToHash("0xeeb86e957f52c5dc4f656cbf4acb70cf96cb52f32494beb79fdcac0ca2e45297")
	data[common.HexToHash("0x37ade56ddeaf07eea736ce3ce05fe8f3a94b058dea597121dca9fe26e0124774")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xbf9d2bb87e62e535ea6ef5ac0a334e29d385528e7e05f1878903c10ffbc7d30c")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0x8f83dea0634358fc7f838269b688e7442ff3056f3b24689e2f9ec5dc24058983")] = common.HexToHash("0x4f1b5f985999fe9da9988a1534f33e16672c7fbecf5d0bc87f8be8dc8a6afe1f")
	data[common.HexToHash("0x2ca1f376d3c6e52201eead87848701a521ddf61d1a66a80535c9b6b96d4c139b")] = common.HexToHash("0x73dd97fa443c86f2b247132a48b50cead897d9fabbd3700d9f2ad22ccf46bc17")
	data[common.HexToHash("0xabcf212daf70cd5ced9a5fab66287bbcbb33f4bd560d5170e5524a2b021a863f")] = common.HexToHash("0x00000000000000000000000000000000534b6ebf40c104b84c2ae630ee2b7324")
	data[common.HexToHash("0x2ca1f376d3c6e52201eead87848701a521ddf61d1a66a80535c9b6b96d4c139c")] = common.HexToHash("0xc29a080f0770f0240963c5746d48bf073a424e4fbb2225021982cc604e2634cd")
	data[common.HexToHash("0xc7e097067f482fae2b5fa795ab3c7b473803ab44b417e1e8f592b4cf18b5f879")] = common.HexToHash("0x0000000000000000000000000000000009a2e1e78b97f39a34d6a3fafeaff37b")
	data[common.HexToHash("0xb040b6e66cd81670d00989b19acb0d885ee1d63ad05e7a61cc496e57333d5359")] = common.HexToHash("0x59364f156ff4043ebf9e07ddde6d628279f1cde2f34906895d15faabbe9fb6bb")
	data[common.HexToHash("0xc3db16296a126ae9001348cdfd0132769dd804474ebd360dbcc20bbc95cc0e6d")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xa2d5f1809db342eae7d1355030756f064b23d2fe9e1a61582355eb60ded82924")] = common.HexToHash("0x0000000000000000000000002584841f3b2ba8781f56b9c59a527f1246afdc17")
	data[common.HexToHash("0xf770100495b2a545e20ee1f9a5be262b5a65491436bbfe155ae8605f066f64a2")] = common.HexToHash("0xb7982598f229847dce21cc0b6c0c5758d883c2d25d0c94aaab5675019a2e00ba")
	data[common.HexToHash("0xb040b6e66cd81670d00989b19acb0d885ee1d63ad05e7a61cc496e57333d535a")] = common.HexToHash("0x18ed48ab621704542e59e91b6d577ced0e983f4b2214db89886ae54b9184f542")

	return GenesisAccount{
		Balance: big.NewInt(0),
		Nonce:   0,
		Storage: data,
		Code:    hexutil.MustDecode("0x6080604052600436106100f9576000357c01000000000000000000000000000000000000000000000000000000009004806373b150981161009c578063a737b18611610076578063a737b18614610b5a578063c1292cc314610b6f578063e91431f714610b84578063ffdd5cf114610b99576100f9565b806373b1509814610ada57806378583f2314610aef5780639382255714610b45576100f9565b8063251c22d1116100d8578063251c22d1146109c85780632f92673214610a6e57806331deb7e114610a9357806372b507e714610aa8576100f9565b8062b54ea6146108e157806316e7f1711461090857806319fe9a3b14610950575b34801561010557600080fd5b503360009081526004602052604090205460c060020a02600160c060020a031981161561039657600160c060020a03198116600090815260036020526040902060060154151561027a57600160c060020a03198082166000908152600360205260408120600160069091018190556002805490910190555468010000000000000000900460c060020a0216156101de576000805468010000000000000000900460c060020a908102600160c060020a0319168252600360205260409091206002018054600160c060020a03168284049092029190911790555b60008054600160c060020a031983168252600360205260408220600201805460c060020a680100000000000000009384900481028190047001000000000000000000000000000000000277ffffffffffffffff000000000000000000000000000000001990921691909117600160c060020a031690915582549084049091026fffffffffffffffff00000000000000001990911617905561031b565b600160c060020a03198116600090815260036020526040812060050154111561031b57600160c060020a0319811660009081526003602052604090206005015443036103208111156102ec57600160c060020a0319821660009081526003602052604090206001600690910155610319565b600160c060020a031982166000908152600360205260409020600681018054830190556007018054820190555b505b600160c060020a0319811660009081526003602052604090204360058201556002015461036190700100000000000000000000000000000000900460c060020a02610bf7565b600160c060020a031981166000908152600360205260409020600201546103919060c060020a9081900402610bf7565b6108de565b3360009081526005602052604081205411156108de57336000908152600560205260409020805460001981019190829081106103ce57fe5b6000918252602080832060048304015460039283166008026101000a900460c060020a02600160c060020a03198116808552929091526040909220549193501580159061041a57508015155b151561042557600080fd5b61042e83610c64565b6104366114de565b61043e6114f9565b828252600160c060020a031985166000908152600360209081526040822060010154818501529082906080908590600b600019f1151561047d57600080fd5b8051600160a060020a0381166000908152600460209081526040808320805467ffffffffffffffff19169055600160c060020a0319898116845260039092529091206002015460c060020a8082029268010000000000000000909204029082161561052957600160c060020a03198216600090815260036020526040902060020180546fffffffffffffffff000000000000000019166801000000000000000060c060020a8404021790555b600160c060020a031981161561057157600160c060020a031981166000908152600360205260409020600201805467ffffffffffffffff191660c060020a840417905561058b565b6000805467ffffffffffffffff191660c060020a84041790555b600080600360008b600160c060020a031916600160c060020a031916815260200190815260200160002060040154119050610160604051908101604052806000600102815260200160006001028152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a03191681526020016000600160a060020a031681526020016000815260200160008152602001600081526020016000815250600360008b600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001556020820151816001015560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160020160106101000a81548167ffffffffffffffff021916908360c060020a9004021790555060a08201518160020160186101000a81548167ffffffffffffffff021916908360c060020a9004021790555060c08201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060e08201518160040155610100820151816005015561012082015181600601556101408201518160070155905050600060c060020a026005600033600160a060020a0316600160a060020a03168152602001908152602001600020898154811015156107e357fe5b90600052602060002090600491828204019190066008026101000a81548167ffffffffffffffff021916908360c060020a90040217905550876005600033600160a060020a0316600160a060020a031681526020019081526020016000208161084c9190611518565b506001805460001901905560408051600160c060020a03198b16815233602082015281517f86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd929181900390910190a180156108d557604051339060009069021e0c0013070adc00009082818181858883f193505050501580156108d3573d6000803e3d6000fd5b505b50505050505050505b50005b3480156108ed57600080fd5b506108f6610df5565b60408051918252519081900360200190f35b34801561091457600080fd5b5061093c6004803603602081101561092b57600080fd5b5035600160c060020a031916610dfb565b604080519115158252519081900360200190f35b34801561095c57600080fd5b506109896004803603604081101561097357600080fd5b50600160a060020a038135169060200135610e19565b604051828152602081018260a080838360005b838110156109b457818101518382015260200161099c565b505050509050019250505060405180910390f35b3480156109d457600080fd5b506109fc600480360360208110156109eb57600080fd5b5035600160c060020a031916610f24565b604080519b8c5260208c019a909a52600160c060020a03199889168b8b015296881660608b015294871660808a01529290951660a0880152600160a060020a031660c087015260e086019390935261010085019290925261012084019190915261014083015251908190036101600190f35b610a9160048036036040811015610a8457600080fd5b5080359060200135610fa7565b005b348015610a9f57600080fd5b506108f6610fb6565b610a9160048036036060811015610abe57600080fd5b5080359060208101359060400135600160a060020a0316610fc4565b348015610ae657600080fd5b506108f6611420565b348015610afb57600080fd5b50610b2860048036036040811015610b1257600080fd5b50600160a060020a038135169060200135611426565b60408051600160c060020a03199092168252519081900360200190f35b348015610b5157600080fd5b506108f661146e565b348015610b6657600080fd5b506108f661147a565b348015610b7b57600080fd5b50610b28611480565b348015610b9057600080fd5b50610b2861148c565b348015610ba557600080fd5b50610bcc60048036036020811015610bbc57600080fd5b5035600160a060020a03166114a4565b6040805195865260208601949094528484019290925260608401526080830152519081900360a00190f35b600160c060020a0319811615801590610c285750600160c060020a0319811660009081526003602052604090205415155b15610c6157600160c060020a0319811660009081526003602052604090206005015461032043919091031115610c6157610c6181610c64565b50565b600160c060020a031981166000908152600360205260408120600601541115610c615760028054600019018155600160c060020a0319808316600090815260036020526040812060068101919091559091015460c060020a7001000000000000000000000000000000008204810292918190040290821615610d3f57600160c060020a031982811660009081526003602052604080822060029081018054600160c060020a031660c060020a808804021790559286168252902001805477ffffffffffffffff00000000000000000000000000000000191690555b600160c060020a0319811615610dc357600160c060020a03198181166000908152600360205260408082206002908101805477ffffffffffffffff00000000000000000000000000000000191670010000000000000000000000000000000060c060020a89040217905592861682529020018054600160c060020a03169055610df0565b600080546fffffffffffffffff000000000000000019166801000000000000000060c060020a8504021790555b505050565b60025481565b600160c060020a031916600090815260036020526040902054151590565b6000610e2361154c565b600160a060020a038416600090815260056020908152604091829020805483518184028101840190945280845260609392830182828015610eb357602002820191906000526020600020906000905b82829054906101000a900460c060020a02600160c060020a03191681526020019060080190602082600701049283019260010382029150808411610e725790505b5050835196509293506000925050505b600581108015610ed4575083858201105b15610f1b5781858201815181101515610ee957fe5b60209081029091010151838260058110610eff57fe5b600160c060020a03199092166020929092020152600101610ec3565b50509250929050565b6003602081905260009182526040909120805460018201546002830154938301546004840154600585015460068601546007909601549496939560c060020a8086029668010000000000000000870482029670010000000000000000000000000000000081048302969083900490920294600160a060020a03909216939192908b565b610fb2828233610fc4565b5050565b69021e19e0c9bab240000081565b82600160c060020a0319811615801590610fdd57508215155b80156110005750600160c060020a03198116600090815260036020526040902054155b8015611015575069021e19e0c9bab240000034145b151561102057600080fd5b6110286114de565b6110306114f9565b8582526020808301869052816080846000600b600019f1151561105257600080fd5b8051600160a060020a038116151561106957600080fd5b836004600083600160a060020a0316600160a060020a0316815260200190815260200160002060006101000a81548167ffffffffffffffff021916908360c060020a90040217905550610160604051908101604052808881526020018781526020016000809054906101000a900460c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a0319168152602001600060c060020a02600160c060020a031916815260200186600160a060020a03168152602001438152602001600081526020016000815260200160008152506003600086600160c060020a031916600160c060020a0319168152602001908152602001600020600082015181600001556020820151816001015560408201518160020160006101000a81548167ffffffffffffffff021916908360c060020a9004021790555060608201518160020160086101000a81548167ffffffffffffffff021916908360c060020a9004021790555060808201518160020160106101000a81548167ffffffffffffffff021916908360c060020a9004021790555060a08201518160020160186101000a81548167ffffffffffffffff021916908360c060020a9004021790555060c08201518160030160006101000a815481600160a060020a030219169083600160a060020a0316021790555060e08201518160040155610100820151816005015561012082015181600601556101408201518160070155905050600060c060020a02600160c060020a0319166000809054906101000a900460c060020a02600160c060020a0319161415156113255760008054600160c060020a031960c060020a91820216825260036020526040909120600201805491860468010000000000000000026fffffffffffffffff0000000000000000199092169190911790555b6000805467ffffffffffffffff191660c060020a86049081178255600160a060020a0387811683526005602090815260408085208054600181810183559187529286206004840401805467ffffffffffffffff60039095166008026101000a9485021916939095029290921790935580548101905590519083169190670de0b6b3a76400009082818181858883f193505050501580156113c9573d6000803e3d6000fd5b5060408051600160c060020a031986168152600160a060020a038716602082015281517ff19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389929181900390910190a150505050505050565b60015481565b60056020528160005260406000208181548110151561144157fe5b9060005260206000209060049182820401919006600802915091509054906101000a900460c060020a0281565b670de0b6b3a764000081565b61032081565b60005460c060020a0281565b60005468010000000000000000900460c060020a0281565b600154600254600160a060020a0392909216600090815260056020526040902054670de0b6b3a764000030310493600a4360300204939190565b60408051808201825290600290829080388339509192915050565b6020604051908101604052806001906020820280388339509192915050565b815481835581811115610df0576003016004900481600301600490048360005260206000209182019101610df0919061156b565b60a0604051908101604052806005906020820280388339509192915050565b61158991905b808211156115855760008155600101611571565b5090565b9056fea165627a7a72305820c044eb3e1caeb73630becc79ac14f023ad04a34dcf52035fa5bc472236874b2b0029"),
	}
}
