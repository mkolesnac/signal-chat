package services

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"signal-chat/cmd/server/models"
)

var privateKey = byteArray32()
var identityKeyPair = ecc.CreateKeyPair(privateKey[:])

var TestingAccount, _ = models.NewAccount("123", "test", "signed1")
var TestingIdentityKey = models.NewIdentityKey("123", identityKeyPair.PublicKey().PublicKey())
var TestingSignedPreKey = NewTestingSignedPreKey("123", "signed1", byteArray32())
var TestingPreKey1 = models.NewPreKey("123", "prekey1", byteArray32())
var TestingPreKey2 = models.NewPreKey("123", "prekey2", byteArray32())

func byteArray32() [32]byte {
	var byteArray [32]byte
	for i := range byteArray {
		byteArray[i] = byte(i) // filling with 0, 1, 2, ..., 63
	}
	return byteArray
}

func NewTestingSignedPreKey(accountID, keyID string, publicKey [32]byte) *models.SignedPreKey {
	signature := ecc.CalculateSignature(identityKeyPair.PrivateKey(), publicKey[:])
	return models.NewSignedPreKey(accountID, keyID, publicKey, signature)
}
