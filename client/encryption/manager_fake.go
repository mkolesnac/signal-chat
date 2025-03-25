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

func (s *ManagerFake) CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error) {
	messages := make(map[string][]byte)
	for _, id := range recipientIDs {
		messages[id] = makeArr(64)
	}
	return messages, nil
}

func (s *ManagerFake) ProcessSenderKeyDistributionMessage(groupID, senderID string, encryptedMsg []byte) error {
	return nil
}

func (s *ManagerFake) GroupEncrypt(groupID string, plaintext []byte) (*EncryptedMessage, error) {
	ciphertext := simpleEncrypt(plaintext)
	return &EncryptedMessage{
		Serialized: ciphertext,
		Ciphertext: ciphertext,
		Envelope: &Envelope{
			KeyID:     0,
			Version:   0,
			Iteration: 0,
			Signature: makeArr(64),
		},
	}, nil
}

func (s *ManagerFake) GroupDecrypt(groupID, senderID string, ciphertext []byte) (*DecryptedMessage, error) {
	return &DecryptedMessage{
		Plaintext:  simpleDecrypt(ciphertext),
		Ciphertext: ciphertext,
		Envelope: &Envelope{
			KeyID:     0,
			Version:   0,
			Iteration: 0,
			Signature: makeArr(64),
		},
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

// SimpleEncrypt encrypts a string by rotating each byte by its position
func simpleEncrypt(input []byte) []byte {
	result := make([]byte, len(input))

	for i := 0; i < len(input); i++ {
		result[i] = input[i] + byte(i%10)
	}

	return result
}

// SimpleDecrypt decrypts a string that was encrypted with SimpleEncrypt
func simpleDecrypt(encrypted []byte) []byte {
	result := make([]byte, len(encrypted))

	for i := 0; i < len(encrypted); i++ {
		result[i] = encrypted[i] - byte(i%10)
	}

	return result
}
