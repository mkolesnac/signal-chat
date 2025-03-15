package encryption

import (
	"signal-chat/internal/api"
)

type ManagerStub struct {
	InitializeKeyStoreResult api.KeyBundle
	InitializeKeyStoreError  error
	EncryptResult            []byte
	EncryptError             error
	DecryptResult            []byte
	DecryptError             error
}

func NewManagerStub() *ManagerStub {
	return &ManagerStub{}
}

func (m *ManagerStub) InitializeKeyStore() (api.KeyBundle, error) {
	return m.InitializeKeyStoreResult, m.InitializeKeyStoreError
}

func (m *ManagerStub) Encrypt(plaintext []byte, recipientID string) ([]byte, error) {
	return m.EncryptResult, m.EncryptError
}

func (m *ManagerStub) Decrypt(ciphertext []byte, senderID string) ([]byte, error) {
	return m.DecryptResult, m.DecryptError
}
