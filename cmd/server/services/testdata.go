package services

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/google/uuid"
	"math/rand"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
)

var testingSignedPreKey = newSignedPreKey()

var test = struct {
	account       models.Account
	identityKey   [32]byte
	signedPreKey1 models.SignedPreKey
	signedPreKey2 models.SignedPreKey
	preKey1       models.PreKey
	preKey2       models.PreKey
}{
	account: models.Account{
		SignedPreKeyID: testingSignedPreKey.KeyID,
		ID:             "123",
		Name:           "test User",
		CreatedAt:      "password123",
	},
	identityKey:   [32]byte(randomBytes(32)),
	signedPreKey1: testingSignedPreKey,
	signedPreKey2: newSignedPreKey(),
	preKey1:       newPreKey(),
	preKey2:       newPreKey(),
}

var testAccountPrimKey = models.GetAccountPrimaryKey(test.account.ID)
var testResources = []storage.Resource{
	// Account
	{
		PrimaryKey:     models.GetAccountPrimaryKey(test.account.ID),
		Name:           &test.account.Name,
		SignedPreKeyID: &test.account.SignedPreKeyID,
	},
	// IdentityKey
	{
		PrimaryKey: models.GetIdentityKeyPrimaryKey(testAccountPrimKey),
		PublicKey:  test.identityKey,
	},
	// SignedPreKey1
	{
		PrimaryKey: models.GetSignedPreKeyPrimaryKey(testAccountPrimKey, test.signedPreKey1.KeyID),
		PublicKey:  [32]byte(test.signedPreKey1.PublicKey),
		Signature:  [64]byte(test.signedPreKey1.Signature),
	},
	// SignedPreKey2
	{
		PrimaryKey: models.GetSignedPreKeyPrimaryKey(testAccountPrimKey, test.signedPreKey2.KeyID),
		PublicKey:  [32]byte(test.signedPreKey2.PublicKey),
		Signature:  [64]byte(test.signedPreKey2.Signature),
	},
	// PreKey1
	{
		PrimaryKey: models.GetPreKeyPrimaryKey(testAccountPrimKey, test.preKey1.KeyID),
		PublicKey:  [32]byte(test.preKey1.PublicKey),
	},
	// PreKey2
	{
		PrimaryKey: models.GetPreKeyPrimaryKey(testAccountPrimKey, test.preKey2.KeyID),
		PublicKey:  [32]byte(test.preKey2.PublicKey),
	},
	// Conversation
	{
		PrimaryKey:           models.GetConversationPrimaryKey("abc"),
		LastMessageSnippet:   stringPtr("Lorem ipsum..."),
		LastMessageTimestamp: stringPtr("28.11.2024"),
		SenderID:             stringPtr("123"),
	},
	// Conversation
	{
		PrimaryKey:           models.GetConversationPrimaryKey("edf"),
		LastMessageSnippet:   stringPtr("Dolor sit..."),
		LastMessageTimestamp: stringPtr("30.11.2024"),
		SenderID:             stringPtr("123"),
	},
}

func stringPtr(s string) *string {
	return &s
}

func newPreKey() models.PreKey {
	privateKey := [32]byte(randomBytes(32))
	publicKey := ecc.CreateKeyPair(privateKey[:]).PublicKey().PublicKey()

	return models.PreKey{
		KeyID:     uuid.New().String(),
		PublicKey: publicKey[:],
	}
}

func newSignedPreKey() models.SignedPreKey {
	privateKey := [32]byte(randomBytes(32))
	publicKey := ecc.CreateKeyPair(privateKey[:]).PublicKey().PublicKey()
	signature := ecc.CalculateSignature(ecc.NewDjbECPrivateKey(TestingIdentityPrivateKey), publicKey[:])

	return models.SignedPreKey{
		KeyID:     uuid.New().String(),
		PublicKey: publicKey[:],
		Signature: signature[:],
	}
}

func randomBytes(length int) []byte {
	byteArray := make([]byte, length)
	for i := range byteArray {
		byteArray[i] = byte(rand.Intn(256))
	}
	return byteArray
}
