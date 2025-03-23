package encryption

import (
	"signal-chat/internal/api"
)

type ManagerStub struct {
	InitializeKeyStoreResult api.KeyBundle
	InitializeKeyStoreError  error
	EncryptResult            *EncryptedMessage
	EncryptError             error
	DecryptResult            *DecryptedMessage
	DecryptError             error
}

func NewManagerStub() *ManagerStub {
	return &ManagerStub{}
}

func (m *ManagerStub) InitializeKeyStore() (api.KeyBundle, error) {
	return m.InitializeKeyStoreResult, m.InitializeKeyStoreError
}

func (m *ManagerStub) EncryptionMaterial(otherUserID string) *Material {
	return &Material{
		MessageKeys: MessageKeys{},
	}
}

func (m *ManagerStub) Encrypt(plaintext []byte, recipientID string) (*EncryptedMessage, error) {
	return m.EncryptResult, m.EncryptError
}

func (m *ManagerStub) Decrypt(ciphertext []byte, senderID string) (*DecryptedMessage, error) {
	return m.DecryptResult, m.DecryptError
}
