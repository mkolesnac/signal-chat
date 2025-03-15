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

func (s *ManagerFake) Encrypt(plaintext []byte, recipientID string) ([]byte, error) {
	return plaintext, nil
}

func (s *ManagerFake) Decrypt(ciphertext []byte, senderID string) ([]byte, error) {
	return ciphertext, nil
}

func makeArr(n int) []byte {
	result := make([]byte, n)

	// Fill array with values from 1 to n
	for i := 0; i < n; i++ {
		result[i] = byte(i + 1)
	}

	return result
}
