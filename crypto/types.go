package crypto

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

const (
	CryptoED25519   = "CryptoED25519"
	CryptoSECP256K1 = "CryptoSECP256K1"
)

type Signature struct {
	tmcrypto.PubKey  `json:"pubkey"`
	Signed  []byte   `json:"signed"`
	R       string   `json:"R"`
	S       string   `json:"S"`
	PubPEM  string   `json:"PubPEM"`
}