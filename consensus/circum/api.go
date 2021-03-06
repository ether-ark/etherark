// Copyright 2018 The auc Authors
// This file is part of the auc library.
//
// The auc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The auc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the auc library. If not, see <http://www.gnu.org/licenses/>.

// Package circum implements the proof-of-authority consensus engine.
package circum

import (
	"github.com/ether-ark/etherark/consensus"
)
// API is a user facing RPC API to allow controlling the delegate and voting
// mechanisms of the delegated-proof-of-stake
type API struct {
	chain consensus.ChainReader
	circum  *Circum
}

func (api *API) Test() {
}