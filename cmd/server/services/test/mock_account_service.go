package test

import (
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
)

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(name, pwd string, identityKey [32]byte, signedPrekey models.SignedPreKey, preKeys []models.PreKey) (models.Account, error) {
	args := m.Called(name, pwd, identityKey, signedPrekey, preKeys)
	return args.Get(0).(models.Account), args.Error(1)
}

func (m *MockAccountService) GetAccount(id string) (models.Account, error) {
	args := m.Called(id)
	return args.Get(0).(models.Account), args.Error(1)
}

func (m *MockAccountService) GetSession(acc models.Account) (models.Session, error) {
	args := m.Called(acc)
	return args.Get(0).(models.Session), args.Error(1)
}

func (m *MockAccountService) GetKeyBundle(id string) (models.KeyBundle, error) {
	args := m.Called(id)
	return args.Get(0).(models.KeyBundle), args.Error(1)
}

func (m *MockAccountService) GetPreKeyCount(acc models.Account) (int, error) {
	args := m.Called(acc)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockAccountService) UploadNewPreKeys(acc models.Account, signedPrekey models.SignedPreKey, preKeys []models.PreKey) error {
	args := m.Called(acc, signedPrekey, preKeys)
	return args.Error(0)
}
