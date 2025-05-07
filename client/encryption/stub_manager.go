package encryption

import "signal-chat/internal/apitypes"

type StubManager struct {
	InitializeKeyStoreResult                 apitypes.KeyBundle
	InitializeKeyStoreError                  error
	CreateEncryptionGroupResult              map[string][]byte
	CreateEncryptionGroupError               error
	ProcessSenderKeyDistributionMessageError error
	GroupEncryptResult                       *EncryptedMessage
	GroupEncryptError                        error
	GroupDecryptResult                       *DecryptedMessage
	GroupDecryptError                        error
}

func NewStubManager() *StubManager {
	return &StubManager{}
}

func (m *StubManager) InitializeKeyStore() (apitypes.KeyBundle, error) {
	return m.InitializeKeyStoreResult, m.InitializeKeyStoreError
}

func (m *StubManager) CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error) {
	return m.CreateEncryptionGroupResult, m.CreateEncryptionGroupError
}

func (m *StubManager) ProcessSenderKeyDistributionMessage(groupID string, senderID string, encryptedMsg []byte) error {
	return m.ProcessSenderKeyDistributionMessageError
}

func (m *StubManager) GroupEncrypt(groupID string, plaintext []byte) (*EncryptedMessage, error) {
	return m.GroupEncryptResult, m.GroupEncryptError
}

func (m *StubManager) GroupDecrypt(groupID, senderID string, ciphertext []byte) (*DecryptedMessage, error) {
	return m.GroupDecryptResult, m.GroupDecryptError
}
