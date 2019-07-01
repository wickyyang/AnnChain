package signer

import (
	"errors"

	"github.com/dappledger/AnnChain/modules/common"
	"github.com/dappledger/AnnChain/modules/crypto"
	"github.com/dappledger/AnnChain/modules/rlp"
	"golang.org/x/crypto/sha3"
)

type Signer interface {
	Gen() (string, string)
	PrivToAddress(string) (string, error)
	Sender(common.Hash, []byte) (common.Address, error)
	Signer(common.Hash, string) ([]byte, error)
	Hash(v interface{}) common.Hash
}

type HomesteadSigner struct{}

func (s *HomesteadSigner) Gen() (address string, privKey string) {

	privkey, err := crypto.GenerateKey()

	if err != nil {
		return "", ""
	}
	return crypto.PubkeyToAddress(privkey.PublicKey).Hex(), common.Bytes2Hex(crypto.FromECDSA(privkey))
}

func (s *HomesteadSigner) PrivToAddress(privKey string) (string, error) {

	privBytes := common.Hex2Bytes(privKey)

	privkey := crypto.ToECDSA(privBytes)

	if privkey == nil {
		return "", errors.New("privkey not right")
	}

	return crypto.PubkeyToAddress(privkey.PublicKey).Hex(), nil
}

func (s *HomesteadSigner) Sender(hash common.Hash, sign []byte) (common.Address, error) {

	pub, err := crypto.Ecrecover(hash[:], sign)
	if err != nil {
		return common.Address{}, err
	}

	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}

	var addr common.Address

	copy(addr[:], crypto.Keccak256(pub[1:])[12:])

	return addr, nil

}

func (s *HomesteadSigner) Signer(hash common.Hash, priv string) ([]byte, error) {

	privBytes := common.Hex2Bytes(priv)

	privkey := crypto.ToECDSA(privBytes)
	if privkey == nil {
		return nil, errors.New("invalid public key")
	}
	return crypto.Sign(hash.Bytes(), privkey)
}

func (s *HomesteadSigner) Hash(v interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, v)
	hw.Sum(h[:0])
	return h
}
