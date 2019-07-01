package signer

import (
	"testing"
)

var sgn Signer

var address, privkey string

var msg string = "123456"

func init() {
	sgn = new(HomesteadSigner)
}

func TestGen(t *testing.T) {
	address, privkey = sgn.Gen()
	t.Log(address, privkey)
}

func TestFromPriv(t *testing.T) {

	addr, err := sgn.PrivToAddress(privkey)

	if err != nil {
		panic(err)
	}

	t.Log(addr)
}

func TestSigner(t *testing.T) {

	signByte, err := sgn.Signer(sgn.Hash(msg), privkey)

	if err != nil {
		panic(err)
	}

	addr, err := sgn.Sender(sgn.Hash(msg), signByte)

	if err != nil {
		panic(err)
	}

	t.Log(addr.Hex(), sgn.Hash(msg).Hex())
}
