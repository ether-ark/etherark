// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"math/big"
	"strings"

	ethereum "github.com/ether-ark/etherark"
	"github.com/ether-ark/etherark/accounts/abi"
	"github.com/ether-ark/etherark/accounts/abi/bind"
	"github.com/ether-ark/etherark/common"
	"github.com/ether-ark/etherark/core/types"
	"github.com/ether-ark/etherark/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"}],\"name\":\"register\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"register2\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"join\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"id\",\"type\":\"bytes8\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"quit\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"baseCost\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"countOnlineNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"countTotalNode\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"startPos\",\"type\":\"uint256\"}],\"name\":\"getIds\",\"outputs\":[{\"name\":\"length\",\"type\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes8[5]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"lockedBalance\",\"type\":\"uint256\"},{\"name\":\"releasedReward\",\"type\":\"uint256\"},{\"name\":\"totalNodes\",\"type\":\"uint256\"},{\"name\":\"onlineNodes\",\"type\":\"uint256\"},{\"name\":\"myNodes\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"id\",\"type\":\"bytes8\"}],\"name\":\"has\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"idsOf\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastOnlineId\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minBlockTimeout\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"nodeCost\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes8\"}],\"name\":\"nodes\",\"outputs\":[{\"name\":\"id1\",\"type\":\"bytes32\"},{\"name\":\"id2\",\"type\":\"bytes32\"},{\"name\":\"preId\",\"type\":\"bytes8\"},{\"name\":\"nextId\",\"type\":\"bytes8\"},{\"name\":\"preOnlineId\",\"type\":\"bytes8\"},{\"name\":\"nextOnlineId\",\"type\":\"bytes8\"},{\"name\":\"coinbase\",\"type\":\"address\"},{\"name\":\"blockRegister\",\"type\":\"uint256\"},{\"name\":\"blockLastPing\",\"type\":\"uint256\"},{\"name\":\"blockOnline\",\"type\":\"uint256\"},{\"name\":\"blockOnlineAcc\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
const ContractBin = `0x00`

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// BaseCost is a free data retrieval call binding the contract method 0x93822557.
//
// Solidity: function baseCost() constant returns(uint256)
func (_Contract *ContractCaller) BaseCost(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "baseCost")
	return *ret0, err
}

// BaseCost is a free data retrieval call binding the contract method 0x93822557.
//
// Solidity: function baseCost() constant returns(uint256)
func (_Contract *ContractSession) BaseCost() (*big.Int, error) {
	return _Contract.Contract.BaseCost(&_Contract.CallOpts)
}

// BaseCost is a free data retrieval call binding the contract method 0x93822557.
//
// Solidity: function baseCost() constant returns(uint256)
func (_Contract *ContractCallerSession) BaseCost() (*big.Int, error) {
	return _Contract.Contract.BaseCost(&_Contract.CallOpts)
}

// CountOnlineNode is a free data retrieval call binding the contract method 0x00b54ea6.
//
// Solidity: function countOnlineNode() constant returns(uint256)
func (_Contract *ContractCaller) CountOnlineNode(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "countOnlineNode")
	return *ret0, err
}

// CountOnlineNode is a free data retrieval call binding the contract method 0x00b54ea6.
//
// Solidity: function countOnlineNode() constant returns(uint256)
func (_Contract *ContractSession) CountOnlineNode() (*big.Int, error) {
	return _Contract.Contract.CountOnlineNode(&_Contract.CallOpts)
}

// CountOnlineNode is a free data retrieval call binding the contract method 0x00b54ea6.
//
// Solidity: function countOnlineNode() constant returns(uint256)
func (_Contract *ContractCallerSession) CountOnlineNode() (*big.Int, error) {
	return _Contract.Contract.CountOnlineNode(&_Contract.CallOpts)
}

// CountTotalNode is a free data retrieval call binding the contract method 0x73b15098.
//
// Solidity: function countTotalNode() constant returns(uint256)
func (_Contract *ContractCaller) CountTotalNode(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "countTotalNode")
	return *ret0, err
}

// CountTotalNode is a free data retrieval call binding the contract method 0x73b15098.
//
// Solidity: function countTotalNode() constant returns(uint256)
func (_Contract *ContractSession) CountTotalNode() (*big.Int, error) {
	return _Contract.Contract.CountTotalNode(&_Contract.CallOpts)
}

// CountTotalNode is a free data retrieval call binding the contract method 0x73b15098.
//
// Solidity: function countTotalNode() constant returns(uint256)
func (_Contract *ContractCallerSession) CountTotalNode() (*big.Int, error) {
	return _Contract.Contract.CountTotalNode(&_Contract.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x19fe9a3b.
//
// Solidity: function getIds(address addr, uint256 startPos) constant returns(uint256 length, bytes8[5] data)
func (_Contract *ContractCaller) GetIds(opts *bind.CallOpts, addr common.Address, startPos *big.Int) (struct {
	Length *big.Int
	Data   [5][8]byte
}, error) {
	ret := new(struct {
		Length *big.Int
		Data   [5][8]byte
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getIds", addr, startPos)
	return *ret, err
}

// GetIds is a free data retrieval call binding the contract method 0x19fe9a3b.
//
// Solidity: function getIds(address addr, uint256 startPos) constant returns(uint256 length, bytes8[5] data)
func (_Contract *ContractSession) GetIds(addr common.Address, startPos *big.Int) (struct {
	Length *big.Int
	Data   [5][8]byte
}, error) {
	return _Contract.Contract.GetIds(&_Contract.CallOpts, addr, startPos)
}

// GetIds is a free data retrieval call binding the contract method 0x19fe9a3b.
//
// Solidity: function getIds(address addr, uint256 startPos) constant returns(uint256 length, bytes8[5] data)
func (_Contract *ContractCallerSession) GetIds(addr common.Address, startPos *big.Int) (struct {
	Length *big.Int
	Data   [5][8]byte
}, error) {
	return _Contract.Contract.GetIds(&_Contract.CallOpts, addr, startPos)
}

// GetInfo is a free data retrieval call binding the contract method 0xffdd5cf1.
//
// Solidity: function getInfo(address addr) constant returns(uint256 lockedBalance, uint256 releasedReward, uint256 totalNodes, uint256 onlineNodes, uint256 myNodes)
func (_Contract *ContractCaller) GetInfo(opts *bind.CallOpts, addr common.Address) (struct {
	LockedBalance  *big.Int
	ReleasedReward *big.Int
	TotalNodes     *big.Int
	OnlineNodes    *big.Int
	MyNodes        *big.Int
}, error) {
	ret := new(struct {
		LockedBalance  *big.Int
		ReleasedReward *big.Int
		TotalNodes     *big.Int
		OnlineNodes    *big.Int
		MyNodes        *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getInfo", addr)
	return *ret, err
}

// GetInfo is a free data retrieval call binding the contract method 0xffdd5cf1.
//
// Solidity: function getInfo(address addr) constant returns(uint256 lockedBalance, uint256 releasedReward, uint256 totalNodes, uint256 onlineNodes, uint256 myNodes)
func (_Contract *ContractSession) GetInfo(addr common.Address) (struct {
	LockedBalance  *big.Int
	ReleasedReward *big.Int
	TotalNodes     *big.Int
	OnlineNodes    *big.Int
	MyNodes        *big.Int
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, addr)
}

// GetInfo is a free data retrieval call binding the contract method 0xffdd5cf1.
//
// Solidity: function getInfo(address addr) constant returns(uint256 lockedBalance, uint256 releasedReward, uint256 totalNodes, uint256 onlineNodes, uint256 myNodes)
func (_Contract *ContractCallerSession) GetInfo(addr common.Address) (struct {
	LockedBalance  *big.Int
	ReleasedReward *big.Int
	TotalNodes     *big.Int
	OnlineNodes    *big.Int
	MyNodes        *big.Int
}, error) {
	return _Contract.Contract.GetInfo(&_Contract.CallOpts, addr)
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(bytes8 id) constant returns(bool)
func (_Contract *ContractCaller) Has(opts *bind.CallOpts, id [8]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "has", id)
	return *ret0, err
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(bytes8 id) constant returns(bool)
func (_Contract *ContractSession) Has(id [8]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// Has is a free data retrieval call binding the contract method 0x16e7f171.
//
// Solidity: function has(bytes8 id) constant returns(bool)
func (_Contract *ContractCallerSession) Has(id [8]byte) (bool, error) {
	return _Contract.Contract.Has(&_Contract.CallOpts, id)
}

// IdsOf is a free data retrieval call binding the contract method 0x78583f23.
//
// Solidity: function idsOf(address , uint256 ) constant returns(bytes8)
func (_Contract *ContractCaller) IdsOf(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) ([8]byte, error) {
	var (
		ret0 = new([8]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "idsOf", arg0, arg1)
	return *ret0, err
}

// IdsOf is a free data retrieval call binding the contract method 0x78583f23.
//
// Solidity: function idsOf(address , uint256 ) constant returns(bytes8)
func (_Contract *ContractSession) IdsOf(arg0 common.Address, arg1 *big.Int) ([8]byte, error) {
	return _Contract.Contract.IdsOf(&_Contract.CallOpts, arg0, arg1)
}

// IdsOf is a free data retrieval call binding the contract method 0x78583f23.
//
// Solidity: function idsOf(address , uint256 ) constant returns(bytes8)
func (_Contract *ContractCallerSession) IdsOf(arg0 common.Address, arg1 *big.Int) ([8]byte, error) {
	return _Contract.Contract.IdsOf(&_Contract.CallOpts, arg0, arg1)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractCaller) LastId(opts *bind.CallOpts) ([8]byte, error) {
	var (
		ret0 = new([8]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastId")
	return *ret0, err
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractSession) LastId() ([8]byte, error) {
	return _Contract.Contract.LastId(&_Contract.CallOpts)
}

// LastId is a free data retrieval call binding the contract method 0xc1292cc3.
//
// Solidity: function lastId() constant returns(bytes8)
func (_Contract *ContractCallerSession) LastId() ([8]byte, error) {
	return _Contract.Contract.LastId(&_Contract.CallOpts)
}

// LastOnlineId is a free data retrieval call binding the contract method 0xe91431f7.
//
// Solidity: function lastOnlineId() constant returns(bytes8)
func (_Contract *ContractCaller) LastOnlineId(opts *bind.CallOpts) ([8]byte, error) {
	var (
		ret0 = new([8]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastOnlineId")
	return *ret0, err
}

// LastOnlineId is a free data retrieval call binding the contract method 0xe91431f7.
//
// Solidity: function lastOnlineId() constant returns(bytes8)
func (_Contract *ContractSession) LastOnlineId() ([8]byte, error) {
	return _Contract.Contract.LastOnlineId(&_Contract.CallOpts)
}

// LastOnlineId is a free data retrieval call binding the contract method 0xe91431f7.
//
// Solidity: function lastOnlineId() constant returns(bytes8)
func (_Contract *ContractCallerSession) LastOnlineId() ([8]byte, error) {
	return _Contract.Contract.LastOnlineId(&_Contract.CallOpts)
}

// MinBlockTimeout is a free data retrieval call binding the contract method 0xa737b186.
//
// Solidity: function minBlockTimeout() constant returns(uint256)
func (_Contract *ContractCaller) MinBlockTimeout(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minBlockTimeout")
	return *ret0, err
}

// MinBlockTimeout is a free data retrieval call binding the contract method 0xa737b186.
//
// Solidity: function minBlockTimeout() constant returns(uint256)
func (_Contract *ContractSession) MinBlockTimeout() (*big.Int, error) {
	return _Contract.Contract.MinBlockTimeout(&_Contract.CallOpts)
}

// MinBlockTimeout is a free data retrieval call binding the contract method 0xa737b186.
//
// Solidity: function minBlockTimeout() constant returns(uint256)
func (_Contract *ContractCallerSession) MinBlockTimeout() (*big.Int, error) {
	return _Contract.Contract.MinBlockTimeout(&_Contract.CallOpts)
}

// NodeCost is a free data retrieval call binding the contract method 0x31deb7e1.
//
// Solidity: function nodeCost() constant returns(uint256)
func (_Contract *ContractCaller) NodeCost(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "nodeCost")
	return *ret0, err
}

// NodeCost is a free data retrieval call binding the contract method 0x31deb7e1.
//
// Solidity: function nodeCost() constant returns(uint256)
func (_Contract *ContractSession) NodeCost() (*big.Int, error) {
	return _Contract.Contract.NodeCost(&_Contract.CallOpts)
}

// NodeCost is a free data retrieval call binding the contract method 0x31deb7e1.
//
// Solidity: function nodeCost() constant returns(uint256)
func (_Contract *ContractCallerSession) NodeCost() (*big.Int, error) {
	return _Contract.Contract.NodeCost(&_Contract.CallOpts)
}

// Nodes is a free data retrieval call binding the contract method 0x251c22d1.
//
// Solidity: function nodes(bytes8 ) constant returns(bytes32 id1, bytes32 id2, bytes8 preId, bytes8 nextId, bytes8 preOnlineId, bytes8 nextOnlineId, address coinbase, uint256 blockRegister, uint256 blockLastPing, uint256 blockOnline, uint256 blockOnlineAcc)
func (_Contract *ContractCaller) Nodes(opts *bind.CallOpts, arg0 [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	PreOnlineId    [8]byte
	NextOnlineId   [8]byte
	Coinbase       common.Address
	BlockRegister  *big.Int
	BlockLastPing  *big.Int
	BlockOnline    *big.Int
	BlockOnlineAcc *big.Int
}, error) {
	ret := new(struct {
		Id1            [32]byte
		Id2            [32]byte
		PreId          [8]byte
		NextId         [8]byte
		PreOnlineId    [8]byte
		NextOnlineId   [8]byte
		Coinbase       common.Address
		BlockRegister  *big.Int
		BlockLastPing  *big.Int
		BlockOnline    *big.Int
		BlockOnlineAcc *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "nodes", arg0)
	return *ret, err
}

// Nodes is a free data retrieval call binding the contract method 0x251c22d1.
//
// Solidity: function nodes(bytes8 ) constant returns(bytes32 id1, bytes32 id2, bytes8 preId, bytes8 nextId, bytes8 preOnlineId, bytes8 nextOnlineId, address coinbase, uint256 blockRegister, uint256 blockLastPing, uint256 blockOnline, uint256 blockOnlineAcc)
func (_Contract *ContractSession) Nodes(arg0 [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	PreOnlineId    [8]byte
	NextOnlineId   [8]byte
	Coinbase       common.Address
	BlockRegister  *big.Int
	BlockLastPing  *big.Int
	BlockOnline    *big.Int
	BlockOnlineAcc *big.Int
}, error) {
	return _Contract.Contract.Nodes(&_Contract.CallOpts, arg0)
}

// Nodes is a free data retrieval call binding the contract method 0x251c22d1.
//
// Solidity: function nodes(bytes8 ) constant returns(bytes32 id1, bytes32 id2, bytes8 preId, bytes8 nextId, bytes8 preOnlineId, bytes8 nextOnlineId, address coinbase, uint256 blockRegister, uint256 blockLastPing, uint256 blockOnline, uint256 blockOnlineAcc)
func (_Contract *ContractCallerSession) Nodes(arg0 [8]byte) (struct {
	Id1            [32]byte
	Id2            [32]byte
	PreId          [8]byte
	NextId         [8]byte
	PreOnlineId    [8]byte
	NextOnlineId   [8]byte
	Coinbase       common.Address
	BlockRegister  *big.Int
	BlockLastPing  *big.Int
	BlockOnline    *big.Int
	BlockOnlineAcc *big.Int
}, error) {
	return _Contract.Contract.Nodes(&_Contract.CallOpts, arg0)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 id1, bytes32 id2) returns()
func (_Contract *ContractTransactor) Register(opts *bind.TransactOpts, id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register", id1, id2)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 id1, bytes32 id2) returns()
func (_Contract *ContractSession) Register(id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2)
}

// Register is a paid mutator transaction binding the contract method 0x2f926732.
//
// Solidity: function register(bytes32 id1, bytes32 id2) returns()
func (_Contract *ContractTransactorSession) Register(id1 [32]byte, id2 [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.Register(&_Contract.TransactOpts, id1, id2)
}

// Register2 is a paid mutator transaction binding the contract method 0x72b507e7.
//
// Solidity: function register2(bytes32 id1, bytes32 id2, address owner) returns()
func (_Contract *ContractTransactor) Register2(opts *bind.TransactOpts, id1 [32]byte, id2 [32]byte, owner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "register2", id1, id2, owner)
}

// Register2 is a paid mutator transaction binding the contract method 0x72b507e7.
//
// Solidity: function register2(bytes32 id1, bytes32 id2, address owner) returns()
func (_Contract *ContractSession) Register2(id1 [32]byte, id2 [32]byte, owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Register2(&_Contract.TransactOpts, id1, id2, owner)
}

// Register2 is a paid mutator transaction binding the contract method 0x72b507e7.
//
// Solidity: function register2(bytes32 id1, bytes32 id2, address owner) returns()
func (_Contract *ContractTransactorSession) Register2(id1 [32]byte, id2 [32]byte, owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Register2(&_Contract.TransactOpts, id1, id2, owner)
}

// ContractJoinIterator is returned from FilterJoin and is used to iterate over the raw logs and unpacked data for Join events raised by the Contract contract.
type ContractJoinIterator struct {
	Event *ContractJoin // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractJoinIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractJoin)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractJoin)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractJoinIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractJoinIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractJoin represents a Join event raised by the Contract contract.
type ContractJoin struct {
	Id   [8]byte
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterJoin is a free log retrieval operation binding the contract event 0xf19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389.
//
// Solidity: event join(bytes8 id, address addr)
func (_Contract *ContractFilterer) FilterJoin(opts *bind.FilterOpts) (*ContractJoinIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return &ContractJoinIterator{contract: _Contract.contract, event: "join", logs: logs, sub: sub}, nil
}

// WatchJoin is a free log subscription operation binding the contract event 0xf19f694d42048723a415f5eed7c402ce2c2e5dc0c41580c3f80e220db85ac389.
//
// Solidity: event join(bytes8 id, address addr)
func (_Contract *ContractFilterer) WatchJoin(opts *bind.WatchOpts, sink chan<- *ContractJoin) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "join")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractJoin)
				if err := _Contract.contract.UnpackLog(event, "join", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ContractQuitIterator is returned from FilterQuit and is used to iterate over the raw logs and unpacked data for Quit events raised by the Contract contract.
type ContractQuitIterator struct {
	Event *ContractQuit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractQuitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractQuit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractQuit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractQuitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractQuitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractQuit represents a Quit event raised by the Contract contract.
type ContractQuit struct {
	Id   [8]byte
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterQuit is a free log retrieval operation binding the contract event 0x86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd.
//
// Solidity: event quit(bytes8 id, address addr)
func (_Contract *ContractFilterer) FilterQuit(opts *bind.FilterOpts) (*ContractQuitIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return &ContractQuitIterator{contract: _Contract.contract, event: "quit", logs: logs, sub: sub}, nil
}

// WatchQuit is a free log subscription operation binding the contract event 0x86d1ab9dbf33cb06567fbeb4b47a6a365cf66f632380589591255187f5ca09cd.
//
// Solidity: event quit(bytes8 id, address addr)
func (_Contract *ContractFilterer) WatchQuit(opts *bind.WatchOpts, sink chan<- *ContractQuit) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "quit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractQuit)
				if err := _Contract.contract.UnpackLog(event, "quit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
