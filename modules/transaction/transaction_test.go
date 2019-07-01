package transaction

import (
	"math/big"
	"testing"
	"time"

	"github.com/dappledger/AnnChain/modules/rlp"
	"github.com/dappledger/AnnChain/modules/signer"
)

func TestTransaction(t *testing.T) {

	sng := new(signer.HomesteadSigner)

	address, privkey := sng.Gen()

	tx := newTransaction(address, big.NewInt(time.Now().UnixNano()), []byte("123"))

	t.Log("txHash:", tx.Hash(sng).Hex())

	stx, err := tx.SignTx(sng, privkey)

	if err != nil {
		t.Log("txSign error:", err)
	}

	addr, err := stx.Sender(sng)

	if err != nil {
		t.Log("txSender error:", err)
	}

	t.Log(addr.Hex(), address)

	rlEncode, err := rlp.EncodeToBytes(tx)
	if err != nil {
		panic(err)
	}

	rltx := Transaction{}

	if err := rlp.DecodeBytes(rlEncode, &rltx); err != nil {
		panic(err)
	}

	t.Log(rltx)

}
