package test

import (
	"github.com/stretchr/testify/mock"
	"signal-chat/server/storage"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetItem(pk, sk string) (storage.Resource, error) {
	args := m.Called(pk, sk)
	return args.Get(0).(storage.Resource), args.Error(1)
}

func (m *MockStorage) QueryItems(pk, skPrefix string, queryCondition storage.QueryCondition) ([]storage.Resource, error) {
	args := m.Called(pk, skPrefix, queryCondition)
	return args.Get(0).([]storage.Resource), args.Error(1)
}

func (m *MockStorage) DeleteItem(pk, sk string) error {
	args := m.Called(pk, sk)
	return args.Error(0)
}

func (m *MockStorage) UpdateItem(pk, sk string, updates map[string]interface{}) error {
	args := m.Called(pk, sk, updates)
	return args.Error(0)
}

func (m *MockStorage) WriteItem(resource storage.Resource) error {
	args := m.Called(resource)
	return args.Error(0)
}

func (m *MockStorage) BatchWriteItems(resources []storage.Resource) error {
	args := m.Called(resources)
	return args.Error(0)
}
