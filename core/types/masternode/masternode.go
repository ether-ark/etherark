// Copyright 2015 The go-ethereum Authors
// Copyright 2018 The go-auc Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package masternode

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ether-ark/etherark/accounts/abi/bind"
	"github.com/ether-ark/etherark/common"
	"github.com/ether-ark/etherark/contracts/masternode/contract"
	"github.com/ether-ark/etherark/crypto"
	"github.com/ether-ark/etherark/p2p/discv5"
	"github.com/ether-ark/etherark/p2p/enode"
	"math/big"
	"sort"
	"encoding/binary"
)

type MasternodeData struct {
	Index      int            `json:"index"     gencodec:"required"`
	Id         string         `json:"id"        gencodec:"required"`
	Data       string         `json:"data"      gencodec:"required"`
	Note       string         `json:"note"      gencodec:"required"`
	Account    common.Address `json:"account"`
	PrivateKey string         `json:"privateKey"       gencodec:"required"`
	PublicKey  string         `json:"publicKey"       gencodec:"required"`

	Coinbase       common.Address `json:"coinbase"`
	Status         uint64         `json:"status"`
	BlockEnd       uint64         `json:"blockEnd"`
	BlockRegister  uint64         `json:"blockRegister"`
	BlockLastPing  uint64         `json:"blockLastPing"`
	BlockOnline    uint64         `json:"blockOnline"`
	BlockOnlineAcc uint64         `json:"blockOnlineAcc"`
	Referrer       string         `json:"referrer"`
}

type MasternodeDatas []*MasternodeData

func (s MasternodeDatas) Len() int {
	return len(s)
}

func (s MasternodeDatas) Less(i, j int) bool {
	return s[i].Index < s[j].Index
}

func (s MasternodeDatas) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Masternode struct {
	ENode *enode.Node

	ID          string
	NodeID      discv5.NodeID
	Account     common.Address
	OriginBlock *big.Int

	BlockOnline    *big.Int
	BlockOnlineAcc *big.Int
	BlockLastPing  *big.Int
}

func newMasternode(nodeId discv5.NodeID, account common.Address,
	block, blockOnline, blockOnlineAcc, blockLastPing *big.Int) *Masternode {
	id := GetMasternodeID(nodeId)
	p := &ecdsa.PublicKey{Curve: crypto.S256(), X: new(big.Int), Y: new(big.Int)}
	p.X.SetBytes(nodeId[:32])
	p.Y.SetBytes(nodeId[32:])
	if !p.Curve.IsOnCurve(p.X, p.Y) {
		return &Masternode{}
	}
	node := enode.NewV4(p, nil, 0, 0)
	return &Masternode{
		ENode:          node,
		ID:             id,
		NodeID:         nodeId,
		Account:        account,
		OriginBlock:    block,
		BlockOnline:    blockOnline,
		BlockOnlineAcc: blockOnlineAcc,
		BlockLastPing:  blockLastPing,
	}
}

func (n *Masternode) String() string {
	return fmt.Sprintf("Node: %s\n", n.NodeID.String())
}

func GetIdsByBlockNumber(contract *contract.Contract, blockNumber *big.Int) ([]string, error) {
	if blockNumber == nil {
		blockNumber = new(big.Int)
	}

	ids, err := getOnlineIds(contract, blockNumber)
	if err == nil && len(ids) > 20 {
		return sortIds(ids), nil
	} else if err != nil {
		fmt.Println("getOnlineIds error:", err)
	}

	ids, err = getAllIds(contract, blockNumber)
	if err != nil {
		fmt.Println("getAllIds error:", err)
	}
	return sortIds(ids), err
}

func getOnlineIds(contract *contract.Contract, blockNumber *big.Int) ([]string, error) {
	opts := new(bind.CallOpts)
	opts.BlockNumber = blockNumber
	var (
		lastId [8]byte
		ctx    *MasternodeContext
		ids    []string
	)
	lastId, err := contract.LastOnlineId(opts)
	if err != nil {
		return ids, err
	}
	for lastId != ([8]byte{}) {
		ctx, err = GetMasternodeContext(opts, contract, lastId)
		if err != nil {
			fmt.Println("getOnlineIds1 error:", err)
			break
		}
		lastId = ctx.preOnline
		if new(big.Int).Sub(blockNumber, ctx.Node.BlockLastPing).Cmp(big.NewInt(420)) > 0 {
			continue
		} else if ctx.Node.BlockOnlineAcc.Cmp(big.NewInt(3000)) < 0 {
			continue
		}
		ids = append(ids, ctx.Node.ID)
	}
	if len(ids) > 20 {
		return ids, nil
	}
	lastId, err = contract.LastOnlineId(opts)
	if err != nil {
		return ids, err
	}
	for lastId != ([8]byte{}) {
		ctx, err = GetMasternodeContext(opts, contract, lastId)
		if err != nil {
			fmt.Println("getOnlineIds2 error:", err)
			break
		}
		lastId = ctx.preOnline
		if new(big.Int).Sub(blockNumber, ctx.Node.BlockLastPing).Cmp(big.NewInt(1200)) > 0 {
			continue
		} else if ctx.Node.BlockOnlineAcc.Cmp(big.NewInt(1)) < 0 {
			continue
		}
		ids = append(ids, ctx.Node.ID)
	}
	return ids, nil
}

func getAllIds(contract *contract.Contract, blockNumber *big.Int) ([]string, error) {
	opts := new(bind.CallOpts)
	opts.BlockNumber = blockNumber
	var (
		lastId [8]byte
		ctx    *MasternodeContext
		ids    []string
	)
	lastId, err := contract.LastId(opts)
	if err != nil {
		return ids, err
	}
	for lastId != ([8]byte{}) {
		ctx, err = GetMasternodeContext(opts, contract, lastId)
		if err != nil {
			fmt.Println("getAllIds error:", err)
			break
		}
		lastId = ctx.pre
		ids = append(ids, ctx.Node.ID)
	}
	return ids, nil
}

func GetMasternodeID(ID discv5.NodeID) string {
	return fmt.Sprintf("%x", ID[:8])
}

type MasternodeContext struct {
	Node       *Masternode
	pre        [8]byte
	next       [8]byte
	preOnline  [8]byte
	nextOnline [8]byte
}

func GetMasternodeContext(opts *bind.CallOpts, contract *contract.Contract, id [8]byte) (*MasternodeContext, error) {
	data, err := contract.ContractCaller.Nodes(opts, id)
	if err != nil {
		return &MasternodeContext{}, err
	}
	id2 := append(data.Id1[:], data.Id2[:]...)
	var nodeId discv5.NodeID
	copy(nodeId[:], id2[:])
	node := newMasternode(nodeId, data.Coinbase, data.BlockRegister, data.BlockOnline, data.BlockOnlineAcc, data.BlockLastPing)

	return &MasternodeContext{
		Node:       node,
		pre:        data.PreId,
		next:       data.NextId,
		preOnline:  data.PreOnlineId,
		nextOnline: data.NextOnlineId,
	}, nil
}

type sortableId struct {
	id string
	score uint64
}

type sortableIds []*sortableId

func (p sortableIds) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p sortableIds) Len() int      { return len(p) }
func (p sortableIds) Less(i, j int) bool {
	if p[i].score < p[j].score {
		return false
	} else if p[i].score > p[j].score {
		return true
	} else {
		return p[i].id > p[j].id
	}
}

func sortIds (ids []string) ([]string) {
	scores := calculateScores(ids)
	sortedIds := sortableIds{}
	for i, s := range scores {
		sortedIds = append(sortedIds, &sortableId{id: i, score: s})
	}
	sort.Sort(sortedIds)
	returnIds := []string{}
	for _, node := range sortedIds {
		returnIds = append(returnIds, node.id)
		//fmt.Println("sortIds", node.id, node.score)

	}
	return returnIds
}

func calculateScores(ids []string) (map[string]uint64) {
	list := make(map[string]uint64)
	for i := 0; i < len(ids); i++ {
		id := ids[i]
		list[id] = binary.BigEndian.Uint64(common.Hex2Bytes(id))
	}
	return list
}