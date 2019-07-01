package transaction

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/dappledger/AnnChain/modules/common"
	"github.com/dappledger/AnnChain/modules/rlp"
	"github.com/dappledger/AnnChain/modules/signer"
)

type Transaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	From      string   `json:"from" gencodec:"required"`
	Timestamp *big.Int `json:"timestamp" gencodec:"required"`
	Value     []byte   `json:"value" gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

func NewTransaction(address string, timestamp *big.Int, value []byte) *Transaction {
	return newTransaction(address, timestamp, value)
}

func newTransaction(address string, timestamp *big.Int, value []byte) *Transaction {
	if len(value) > 0 {
		value = common.CopyBytes(value)
	}
	d := txdata{
		From:      address,
		Timestamp: timestamp,
		Value:     value,
		V:         new(big.Int),
		R:         new(big.Int),
		S:         new(big.Int),
	}
	return &Transaction{data: d}
}

func (tx *Transaction) Data() []byte { return common.CopyBytes(tx.data.Value) }

func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}

func (tx *Transaction) From() common.Address {
	return common.HexToAddress(tx.data.From)
}

func (tx *Transaction) Timestamp() *big.Int {
	return tx.data.Timestamp
}

func (tx *Transaction) Value() []byte {
	return tx.data.Value
}

func (tx *Transaction) Hash(sgn signer.Signer) common.Hash {

	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	hash := sgn.Hash([]interface{}{tx.From(), tx.Timestamp(), tx.Value()})

	tx.hash.Store(hash)

	return hash
}

func (tx *Transaction) SignTx(sgn signer.Signer, privKey string) (sngTx *Transaction, err error) {

	signBytes, err := sgn.Signer(tx.Hash(sgn), privKey)

	if err != nil {
		return nil, err
	}

	return tx.withSignature(signBytes)
}

func (tx *Transaction) Sender(sgn signer.Signer) (common.Address, error) {

	signBytes, err := tx.recoverSign()

	if err != nil {
		return common.Address{}, err
	}

	return sgn.Sender(tx.Hash(sgn), signBytes)
}

func (tx *Transaction) withSignature(sig []byte) (*Transaction, error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	v := new(big.Int).SetBytes([]byte{sig[64] + 27})
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

func (tx *Transaction) recoverSign() ([]byte, error) {

	if tx.data.V.BitLen() > 8 {
		return nil, errors.New("invalid transaction v, r, s values")
	}

	V := byte(tx.data.V.Uint64() - 27)

	r, s := tx.data.R.Bytes(), tx.data.S.Bytes()

	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)

	copy(sig[64-len(s):64], s)

	sig[64] = V

	return sig, nil
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}
