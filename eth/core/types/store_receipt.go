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

package types

import (
	"bytes"
	"fmt"
	"io"

	"github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/rlp"
)

type SReceipt struct {
	// Consensus fields
	Height    uint64         `json:"height"`
	Timestamp uint64         `json:"timestamp"`
	From      common.Address `json:"from"`
	Value     []byte         `json:"value"`
	TxHash    common.Hash    `json:"txhash" gencodec:"required"`
	Status    uint64         `json:"status"`
}

// receiptRLP is the consensus encoding of a receipt.
type sreceiptRLP struct {
	Height            uint64
	Timestamp         uint64
	From              common.Address
	Value             []byte
	PostStateOrStatus []byte
	TxHash            common.Hash
}

type sreceiptStorageRLP struct {
	Height            uint64
	Timestamp         uint64
	From              common.Address
	Value             []byte
	PostStateOrStatus []byte
	TxHash            common.Hash
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewSReceipt(root []byte) *Receipt {
	return &Receipt{}
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (r *SReceipt) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &sreceiptRLP{r.Height, r.Timestamp, r.From, r.Value, r.statusEncoding(), r.TxHash})
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (r *SReceipt) DecodeRLP(s *rlp.Stream) error {
	var dec sreceiptRLP
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if err := r.setStatus(dec.PostStateOrStatus); err != nil {
		return err
	}
	r.Height, r.Timestamp, r.From, r.Value, r.TxHash = dec.Height, dec.Timestamp, dec.From, dec.Value, dec.TxHash
	return nil
}

func (r *SReceipt) setStatus(postStateOrStatus []byte) error {
	switch {
	case bytes.Equal(postStateOrStatus, receiptStatusSuccessfulRLP):
		r.Status = ReceiptStatusSuccessful
	case bytes.Equal(postStateOrStatus, receiptStatusFailedRLP):
		r.Status = ReceiptStatusFailed
	default:
		return fmt.Errorf("invalid receipt status %x", postStateOrStatus)
	}
	return nil
}

func (r *SReceipt) statusEncoding() []byte {

	switch r.Status {
	case ReceiptStatusFailed:
		return receiptStatusFailedRLP
	default:
		return receiptStatusSuccessfulRLP
	}

}

// ReceiptForStorage is a wrapper around a Receipt that flattens and parses the
// entire content of a receipt, as opposed to only the consensus fields originally.
type SReceiptForStorage SReceipt

// EncodeRLP implements rlp.Encoder, and flattens all content fields of a receipt
// into an RLP stream.
func (r *SReceiptForStorage) EncodeRLP(w io.Writer) error {
	enc := &sreceiptStorageRLP{
		Height:    r.Height,
		Timestamp: r.Timestamp,
		From:      r.From,
		Value:     r.Value,
		TxHash:    r.TxHash,
	}
	return rlp.Encode(w, enc)
}

// DecodeRLP implements rlp.Decoder, and loads both consensus and implementation
// fields of a receipt from an RLP stream.
func (r *SReceiptForStorage) DecodeRLP(s *rlp.Stream) error {
	var dec sreceiptStorageRLP
	if err := s.Decode(&dec); err != nil {
		return err
	}
	// Assign the implementation fields
	r.Height, r.Timestamp, r.From, r.Value, r.TxHash = dec.Height, dec.Timestamp, dec.From, r.Value, dec.TxHash
	return nil
}

// Receipts is a wrapper around a Receipt array to implement DerivableList.
type SReceipts []*SReceipt

// Len returns the number of receipts in this list.
func (r SReceipts) Len() int { return len(r) }

// GetRlp returns the RLP encoding of one receipt from the list.
func (r SReceipts) GetRlp(i int) []byte {
	bytes, err := rlp.EncodeToBytes(r[i])
	if err != nil {
		panic(err)
	}
	return bytes
}
