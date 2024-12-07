package test

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/google/uuid"
	"signal-chat-server/models"
	"signal-chat-server/storage"
	"signal-chat-server/utils"
)

var testingIdentityPrivateKey = [32]byte(utils.RandomBytes(32))
var testingIdentityPublicKey = ecc.CreateKeyPair(testingIdentityPrivateKey[:]).PublicKey().PublicKey()
var testingSignedPreKey = newSignedPreKey()

var testAccountID = "123"
var ConversationID = "abc"
var timestamp = storage.GetTimestamp()

var Model = struct {
	Account       models.Account
	IdentityKey   [32]byte
	SignedPreKey1 models.SignedPreKey
	SignedPreKey2 models.SignedPreKey
	PreKey1       models.PreKey
	PreKey2       models.PreKey
}{
	Account: models.Account{
		SignedPreKeyID: testingSignedPreKey.KeyID,
		ID:             testAccountID,
		Name:           "Test User",
		CreatedAt:      timestamp,
		PasswordHash:   utils.RandomBytes(64),
	},
	IdentityKey:   testingIdentityPublicKey,
	SignedPreKey1: testingSignedPreKey,
	SignedPreKey2: newSignedPreKey(),
	PreKey1:       newPreKey(),
	PreKey2:       newPreKey(),
}

var testAccountPrimKey = models.AccountPrimaryKey(Model.Account.ID)

var Resource = struct {
	Account          storage.Resource
	IdentityKey      storage.Resource
	SignedPreKey1    storage.Resource
	SignedPreKey2    storage.Resource
	PreKey1          storage.Resource
	PreKey2          storage.Resource
	ConversationMeta storage.Resource
	Participant      storage.Resource
	Message          storage.Resource
}{
	Account: storage.Resource{
		PrimaryKey:     testAccountPrimKey,
		Name:           StringPtr("Test User"),
		SignedPreKeyID: &testingSignedPreKey.KeyID,
		CreatedAt:      timestamp,
		PasswordHash:   Model.Account.PasswordHash,
	},
	IdentityKey: storage.Resource{
		PrimaryKey: models.IdentityKeyPrimaryKey(Model.Account.ID),
		PublicKey:  testingIdentityPublicKey,
		CreatedAt:  timestamp,
	},
	SignedPreKey1: storage.Resource{
		PrimaryKey: models.SignedPreKeyPrimaryKey(Model.Account.ID, Model.SignedPreKey1.KeyID),
		PublicKey:  [32]byte(Model.SignedPreKey1.PublicKey),
		Signature:  [64]byte(Model.SignedPreKey1.Signature),
		CreatedAt:  timestamp,
	},
	SignedPreKey2: storage.Resource{
		PrimaryKey: models.SignedPreKeyPrimaryKey(Model.Account.ID, Model.SignedPreKey2.KeyID),
		PublicKey:  [32]byte(Model.SignedPreKey2.PublicKey),
		Signature:  [64]byte(Model.SignedPreKey2.Signature),
		CreatedAt:  timestamp,
	},
	PreKey1: storage.Resource{
		PrimaryKey: models.PreKeyPrimaryKey(Model.Account.ID, Model.PreKey1.KeyID),
		PublicKey:  [32]byte(Model.PreKey1.PublicKey),
		CreatedAt:  timestamp,
	},
	PreKey2: storage.Resource{
		PrimaryKey: models.PreKeyPrimaryKey(Model.Account.ID, Model.PreKey2.KeyID),
		PublicKey:  [32]byte(Model.PreKey2.PublicKey),
		CreatedAt:  timestamp,
	},
	ConversationMeta: storage.Resource{
		PrimaryKey:           models.ConversationMetaPrimaryKey(Model.Account.ID, ConversationID),
		LastMessageSnippet:   StringPtr("asdgsdgsdgsdg"),
		LastMessageTimestamp: &timestamp,
		SenderID:             &Model.Account.ID,
		CreatedAt:            timestamp,
	},
	Participant: storage.Resource{
		PrimaryKey: models.ParticipantPrimaryKey(ConversationID, Model.Account.ID),
		Name:       &Model.Account.Name,
		CreatedAt:  timestamp,
	},
	Message: storage.Resource{
		PrimaryKey: models.MessagePrimaryKey(ConversationID, uuid.New().String()),
		CipherText: StringPtr("asdgsdgsdgsdg"),
		CreatedAt:  timestamp,
		SenderID:   &Model.Account.ID,
	},
}

var Resources = []storage.Resource{Resource.Account, Resource.IdentityKey, Resource.SignedPreKey1, Resource.SignedPreKey2, Resource.PreKey1, Resource.PreKey2, Resource.ConversationMeta, Resource.Participant, Resource.Message}

func StringPtr(s string) *string {
	return &s
}

func newPreKey() models.PreKey {
	privateKey := [32]byte(utils.RandomBytes(32))
	publicKey := ecc.CreateKeyPair(privateKey[:]).PublicKey().PublicKey()

	return models.PreKey{
		KeyID:     uuid.New().String(),
		PublicKey: publicKey[:],
	}
}

func newSignedPreKey() models.SignedPreKey {
	privateKey := [32]byte(utils.RandomBytes(32))
	publicKey := ecc.CreateKeyPair(privateKey[:]).PublicKey().PublicKey()
	signature := ecc.CalculateSignature(ecc.NewDjbECPrivateKey(testingIdentityPrivateKey), publicKey[:])

	return models.SignedPreKey{
		KeyID:     uuid.New().String(),
		PublicKey: publicKey[:],
		Signature: signature[:],
	}
}
