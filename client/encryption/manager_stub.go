package encryption

import (
	"signal-chat/internal/api"
)

type ManagerStub struct {
	InitializeKeyStoreResult                 api.KeyBundle
	InitializeKeyStoreError                  error
	CreateEncryptionGroupResult              map[string][]byte
	CreateEncryptionGroupError               error
	ProcessSenderKeyDistributionMessageError error
	GroupEncryptResult                       *EncryptedMessage
	GroupEncryptError                        error
	GroupDecryptResult                       *DecryptedMessage
	GroupDecryptError                        error
}

func NewManagerStub() *ManagerStub {
	return &ManagerStub{}
}

func (m *ManagerStub) InitializeKeyStore() (api.KeyBundle, error) {
	return m.InitializeKeyStoreResult, m.InitializeKeyStoreError
}

func (m *ManagerStub) CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error) {
	return m.CreateEncryptionGroupResult, m.CreateEncryptionGroupError
}

func (m *ManagerStub) ProcessSenderKeyDistributionMessage(groupID string, senderID string, encryptedMsg []byte) error {
	return m.ProcessSenderKeyDistributionMessageError
}

func (m *ManagerStub) GroupEncrypt(groupID string, plaintext []byte) (*EncryptedMessage, error) {
	return m.GroupEncryptResult, m.GroupEncryptError
}

func (m *ManagerStub) GroupDecrypt(groupID, senderID string, ciphertext []byte) (*DecryptedMessage, error) {
	return m.GroupDecryptResult, m.GroupDecryptError
}
