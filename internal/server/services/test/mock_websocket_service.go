package test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

type MockWebsocketManager struct {
	mock.Mock
}

func (m *MockWebsocketManager) RegisterClient(id string, conn *websocket.Conn) {
	m.Called(id, conn)
}

func (m *MockWebsocketManager) UnregisterClient(id string) {
	m.Called(id)
}

func (m *MockWebsocketManager) SendToClient(receiverID string, msg interface{}) error {
	args := m.Called(receiverID, msg)
	return args.Error(0)
}
