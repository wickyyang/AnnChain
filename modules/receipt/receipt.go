// Copyright 2014 The go-ethereum Authors
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

package receipt

import (
	"github.com/dappledger/AnnChain/modules/common"
	"github.com/dappledger/AnnChain/modules/rlp"
)

type Receipt struct {
	Height    uint64         `json:"height"`
	Timestamp uint64         `json:"timestamp"`
	From      common.Address `json:"from"`
	Value     []byte         `json:"value"`
	Op        byte           `json:"opcode"`
	TxHash    common.Hash    `json:"txhash" gencodec:"required"`
}

func (r *Receipt) EncodeRlp() ([]byte, error) {
	return rlp.EncodeToBytes(r)
}

type Receipts []*Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
