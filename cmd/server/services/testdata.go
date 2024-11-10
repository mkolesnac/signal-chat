package services

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"signal-chat/cmd/server/models"
)

var privateKey = byteArray32()
var identityKeyPair = ecc.CreateKeyPair(privateKey[:])

var TestingAccount, _ = models.NewAccount("Test User", "test", "signed1")
var TestingIdentityKey = models.NewIdentityKey(TestingAccount.GetID(), identityKeyPair.PublicKey().PublicKey())
var TestingSignedPreKey = NewTestingSignedPreKey(TestingAccount.GetID(), "signed1", byteArray32())
var TestingPreKey1 = models.NewPreKey(TestingAccount.GetID(), "prekey1", byteArray32())
var TestingPreKey2 = models.NewPreKey(TestingAccount.GetID(), "prekey2", byteArray32())

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
