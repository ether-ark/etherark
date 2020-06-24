// Copyright 2018 The go-auc Authors
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

package eth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"crypto/ecdsa"
	"github.com/ether-ark/etherark"
	"github.com/ether-ark/etherark/common"
	"github.com/ether-ark/etherark/common/math"
	"github.com/ether-ark/etherark/contracts/masternode/contract"
	"github.com/ether-ark/etherark/core/types"
	"github.com/ether-ark/etherark/core/types/masternode"
	"github.com/ether-ark/etherark/crypto"
	"github.com/ether-ark/etherark/eth/downloader"
	"github.com/ether-ark/etherark/event"
	"github.com/ether-ark/etherark/log"
	"github.com/ether-ark/etherark/p2p"
	"github.com/ether-ark/etherark/params"
)

var (
	ErrUnknownMasternode = errors.New("unknown masternode")
)

type x8 [8]byte

type MasternodeManager struct {
	srvr     *p2p.Server
	contract *contract.Contract

	mux *event.TypeMux
	eth *Ethereum

	isMasternode uint32
	syncing      uint32

	coinbase  common.Address
	referrers []common.Address

	mu sync.RWMutex
	rw sync.RWMutex

	ID          string
	id8         x8
	NodeAccount common.Address
	PrivateKey  *ecdsa.PrivateKey
}

func NewMasternodeManager(eth *Ethereum) (*MasternodeManager, error) {
	contractBackend := NewContractBackend(eth)
	contract, err := contract.NewContract(params.MasterndeContractAddress, contractBackend)
	if err != nil {
		return nil, err
	}
	// Create the masternode manager with its initial settings
	manager := &MasternodeManager{
		eth:          eth,
		contract:     contract,
		syncing:      0,
		isMasternode: 0,
	}
	return manager, nil
}

func (self *MasternodeManager) IsMasternode() bool {
	return atomic.LoadUint32(&self.isMasternode) == 1
}

func (self *MasternodeManager) Start(srvr *p2p.Server, mux *event.TypeMux) {
	self.mux = mux
	log.Info("MasternodeManqager Start ...")
	self.PrivateKey = srvr.Config.PrivateKey
	self.NodeAccount = crypto.PubkeyToAddress(self.PrivateKey.PublicKey)
	self.srvr = srvr
	self.id8 = self.X8(self.PrivateKey)
	self.ID = self.fromX8(self.id8)
	self.activeMasternode(self.id8)

	if atomic.LoadUint32(&self.isMasternode) == 1 {
		fmt.Printf("### Conbase: %s\n", self.coinbase.String())
	}
	go self.masternodeLoop()
	go self.checkSyncing()
}

func (self *MasternodeManager) checkSyncing() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			atomic.StoreUint32(&self.syncing, 1)
		case downloader.DoneEvent, downloader.FailedEvent:
			atomic.StoreUint32(&self.syncing, 0)
		}
	}
}

func (self *MasternodeManager) CheckMasternodeId(id string) bool {
	if id == self.ID {
		return true
	}
	return false
}

func (self *MasternodeManager) MasternodeList(number *big.Int) ([]string, error) {
	return masternode.GetIdsByBlockNumber(self.contract, number)
}

func (self *MasternodeManager) GetRefAddr() (common.Address, []common.Address) {
	return self.coinbase, self.referrers
}

func (self *MasternodeManager) SignHash(id string, hash []byte) ([]byte, error) {
	// Look up the key to sign with and abort if it cannot be found
	self.rw.RLock()
	defer self.rw.RUnlock()

	if self.CheckMasternodeId(id) {
		// Sign the hash using plain ECDSA operations
		return crypto.Sign(hash, self.PrivateKey)
	}

	return nil, ErrUnknownMasternode
}

// X8 returns 8 bytes of ecdsa.PublicKey.X
func (self *MasternodeManager) X8(key *ecdsa.PrivateKey) (id x8) {
	buf := make([]byte, 32)
	math.ReadBits(key.PublicKey.X, buf)
	copy(id[:], buf[:8])
	return id
}

func (self *MasternodeManager) fromX8(id x8) string {
	return fmt.Sprintf("%x", id[:])
}

func (self *MasternodeManager) XY(key *ecdsa.PrivateKey) (xy [64]byte) {
	pubkey := key.PublicKey
	math.ReadBits(pubkey.X, xy[:32])
	math.ReadBits(pubkey.Y, xy[32:])
	return xy
}

func (self *MasternodeManager) masternodeLoop() {
	joinCh := make(chan *contract.ContractJoin, 32)
	quitCh := make(chan *contract.ContractQuit, 32)
	joinSub, err1 := self.contract.WatchJoin(nil, joinCh)
	if err1 != nil {
		// TODO: exit
		return
	}
	quitSub, err2 := self.contract.WatchQuit(nil, quitCh)
	if err2 != nil {
		// TODO: exit
		return
	}

	ping := time.NewTimer(10 * time.Minute)
	defer ping.Stop()
	//ntp := time.NewTimer(60 * time.Second)
	//defer ntp.Stop()

	for {
		select {
		case err := <- joinSub.Err():
			joinSub.Unsubscribe()
			fmt.Println("eventJoin err", err.Error())
		case err := <-quitSub.Err():
			quitSub.Unsubscribe()
			fmt.Println("eventQuit err", err.Error())
		case join := <-joinCh:
			if self.CheckMasternodeId(self.fromX8(join.Id)) {
				self.activeMasternode(join.Id)
			}
		case quit := <-quitCh:
			if self.CheckMasternodeId(self.fromX8(quit.Id)) {
				fmt.Printf("### [%x] Remove masternode! \n", quit.Id)
				atomic.StoreUint32(&self.isMasternode, 0)
			}
		//case <-ntp.C:
		//	ntp.Reset(10 * time.Minute)
		//	go discover.CheckClockDrift()
		case <-ping.C:
			logTime := time.Now().Format("[2006-01-02 15:04:05]")
			ping.Reset(20 * time.Minute)
			if atomic.LoadUint32(&self.syncing) == 1 {
				fmt.Println(logTime, " syncing...")
				break
			}
			stateDB, _ := self.eth.blockchain.State()
			contractBackend := NewContractBackend(self.eth)
			if atomic.LoadUint32(&self.isMasternode) == 1 {
				address := self.NodeAccount
				if stateDB.GetBalance(address).Cmp(big.NewInt(1e+18)) < 0 {
					fmt.Println(logTime, "Expect to deposit 1 AUC to ", address.String())
					continue
				}
				gasPrice, err := self.eth.APIBackend.gpo.SuggestPrice(context.Background())
				if err != nil {
					fmt.Println("Get gas price error:", err)
					gasPrice = big.NewInt(10e+9)
				}
				msg := ethereum.CallMsg{From: address, To: &params.MasterndeContractAddress}
				gas, err := contractBackend.EstimateGas(context.Background(), msg)
				if err != nil {
					fmt.Println("Get gas error:", err)
					continue
				}
				minPower := new(big.Int).Mul(big.NewInt(int64(gas)), gasPrice)
				// fmt.Println("Gas:", gas, "GasPrice:", gasPrice.String(), "minPower:", minPower.String())
				if stateDB.GetPower(address, self.eth.blockchain.CurrentBlock().Number()).Cmp(minPower) < 0 {
					fmt.Println(logTime, "Insufficient power for ping transaction.", address.Hex(), self.eth.blockchain.CurrentBlock().Number().String(), stateDB.GetPower(address, self.eth.blockchain.CurrentBlock().Number()).String())
					continue
				}
				tx := types.NewTransaction(
					self.eth.txPool.State().GetNonce(address),
					params.MasterndeContractAddress,
					big.NewInt(0),
					gas,
					gasPrice,
					nil,
				)
				signed, err := types.SignTx(tx, types.NewEIP155Signer(self.eth.blockchain.Config().ChainID), self.PrivateKey)
				if err != nil {
					fmt.Println(logTime, "SignTx error:", err)
					continue
				}
				if err := self.eth.txPool.AddLocal(signed); err != nil {
					fmt.Println(logTime, "send ping to txpool error:", err)
					continue
				}
				fmt.Printf("%s Announcement! (%s)\n", logTime, address.String())

				if !self.eth.IsMining() {
					self.eth.StartMining(0)
				}
			} else {
				self.activeMasternode(self.id8)
			}
		}
	}
}

func (self *MasternodeManager) activeMasternode(id8 x8) {
	node, err := self.contract.Nodes(nil, id8)
	if err != nil {
		fmt.Println("[MN] activeMasternode Error:", err)
		return
	}
	if self.coinbase == (common.Address{}) && node.Coinbase != (common.Address{}) {
		self.coinbase = node.Coinbase
		atomic.StoreUint32(&self.isMasternode, 1)
		fmt.Printf("### Active masternode! (%x)\n", id8)
	}
}
