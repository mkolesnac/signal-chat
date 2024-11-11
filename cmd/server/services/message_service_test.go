package services

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services/mocks"
	"signal-chat/cmd/server/storage"
	"signal-chat/internal/api"
	"testing"
)

var TestMessage1 = models.Message{
	TableItem:  storage.TableItem{PartitionKey: "pk1", SortKey: "sk1"},
	SenderID:   "xxx",
	CipherText: "123",
}
var TestMessage2 = models.Message{
	TableItem:  storage.TableItem{PartitionKey: "pk1", SortKey: "sk2"},
	SenderID:   "xxx",
	CipherText: "123",
}

func TestMessageService_GetMessages(t *testing.T) {
	t.Run("when since timestamp", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		mockWebsockets := new(mocks.MockWebsocketManager)
		service := NewMessageService(mockStorage, mockAccounts, mockWebsockets)
		// setup mocks
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, storage.QUERY_GREATER_THAN, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(3).(*[]models.Message) = []models.Message{TestMessage1, TestMessage2}
			}).
			Return(nil)

		// Act
		messages, err := service.GetMessages("acc1", 1, "")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, messages, 2)
	})
}

func TestMessageService_SendMessage(t *testing.T) {
	t.Run("returns message ID on success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		mockWebsockets := new(mocks.MockWebsocketManager)
		service := NewMessageService(mockStorage, mockAccounts, mockWebsockets)
		// setup mocks
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Message")).Return(nil)
		mockWebsockets.On("SendToClient", TestingAccount.GetID(), mock.AnythingOfType("*models.Message")).Return(nil)

		// Act
		req := api.SendMessageRequest{CipherText: "abc123"}
		id, err := service.SendMessage("sender1", TestingAccount.GetID(), req) // send message to TestingAccount

		// Assert
		mockAccounts.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockWebsockets.AssertExpectations(t)
		assert.NoError(t, err)
		assert.NotNil(t, id)
	})
	t.Run("returns error when recipient account doesn't exist", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		mockWebsockets := new(mocks.MockWebsocketManager)
		service := NewMessageService(mockStorage, mockAccounts, mockWebsockets)
		// setup mocks
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(nil, ErrAccountNotFound)

		// Act
		req := api.SendMessageRequest{CipherText: "abc123"}
		_, err := service.SendMessage("sender1", TestingAccount.GetID(), req) // send message to TestingAccount

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotFound)
	})
	t.Run("returns error when write fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		mockWebsockets := new(mocks.MockWebsocketManager)
		service := NewMessageService(mockStorage, mockAccounts, mockWebsockets)
		// setup mocks
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Message")).Return(errors.New("write failed"))

		// Act
		req := api.SendMessageRequest{CipherText: "abc123"}
		_, err := service.SendMessage("sender1", TestingAccount.GetID(), req) // send message to TestingAccount

		// Assert
		assert.Error(t, err)
		mockWebsockets.AssertNotCalled(t, "SendToClient")
	})
	t.Run("ignores websockets failure", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		mockWebsockets := new(mocks.MockWebsocketManager)
		service := NewMessageService(mockStorage, mockAccounts, mockWebsockets)
		// setup mocks
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Message")).Return(nil)
		mockWebsockets.On("SendToClient", TestingAccount.GetID(), mock.AnythingOfType("*models.Message")).Return(errors.New("send failed"))

		// Act
		req := api.SendMessageRequest{CipherText: "abc123"}
		id, err := service.SendMessage("sender1", TestingAccount.GetID(), req) // send message to TestingAccount

		// Assert
		mockAccounts.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockWebsockets.AssertExpectations(t)
		assert.NoError(t, err)
		assert.NotNil(t, id)
	})
}
