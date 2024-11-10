package mocks

import (
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
	"signal-chat/internal/api"
)

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(name, pwd string, req api.CreateAccountRequest) (string, error) {
	args := m.Called(name, pwd, req)
	return args.String(0), args.Error(1)
}

func (m *MockAccountService) GetAccount(id string) (*models.Account, error) {
	args := m.Called(id)
	account, _ := args.Get(0).(*models.Account)
	return account, args.Error(1)
}
