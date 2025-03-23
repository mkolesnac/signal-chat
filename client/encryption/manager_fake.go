package encryption

import (
	"signal-chat/internal/api"
)

type ManagerFake struct {
}

func NewManagerFake() *ManagerFake {
	return &ManagerFake{}
}

func (s *ManagerFake) InitializeKeyStore() (api.KeyBundle, error) {
	return api.KeyBundle{
		RegistrationID: 0,
		IdentityKey:    makeArr(32),
		SignedPreKey: api.SignedPreKey{
			ID:        0,
			PublicKey: makeArr(32),
			Signature: makeArr(64),
		},
		PreKeys: []api.PreKey{{
			ID:        0,
			PublicKey: makeArr(32),
		}},
	}, nil
}

func (s *ManagerFake) EncryptionMaterial(otherUserID string) *Material {
	return &Material{
		MessageKeys: MessageKeys{},
	}
}

func (s *ManagerFake) Encrypt(plaintext []byte, recipientID string) (*EncryptedMessage, error) {
	return &EncryptedMessage{
		Serialized: plaintext,
		Ciphertext: plaintext,
		Envelope:   &Envelope{},
	}, nil
}

func (s *ManagerFake) Decrypt(ciphertext []byte, senderID string) (*DecryptedMessage, error) {
	return &DecryptedMessage{
		Plaintext:  ciphertext,
		Ciphertext: ciphertext,
		Envelope:   &Envelope{},
	}, nil
}

func makeArr(n int) []byte {
	result := make([]byte, n)

	// Fill array with values from 1 to n
	for i := 0; i < n; i++ {
		result[i] = byte(i + 1)
	}

	return result
}
