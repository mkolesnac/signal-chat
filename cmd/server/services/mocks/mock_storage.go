package mocks

import (
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/storage"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetItem(pk, sk string, outPtr any) error {
	args := m.Called(pk, sk, outPtr)
	return args.Error(0)
}

func (m *MockStorage) DeleteItem(pk, sk string) error {
	args := m.Called(pk, sk)
	return args.Error(0)
}

func (m *MockStorage) WriteItem(item storage.PrimaryKeyProvider) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockStorage) BatchWriteItems(items []storage.PrimaryKeyProvider) error {
	args := m.Called(items)
	return args.Error(0)
}

func (m *MockStorage) QueryItems(pk, skPrefix string, queryCondition storage.QueryCondition, out interface{}) error {
	args := m.Called(pk, skPrefix, queryCondition, out)
	return args.Error(0)
}

func (m *MockStorage) QueryItemsBySenderID(senderID, skPrefix string, queryCondition storage.QueryCondition, out interface{}) error {
	args := m.Called(senderID, skPrefix, queryCondition, out)
	return args.Error(0)
}
