package services

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services/test"
	"signal-chat/cmd/server/storage"
	"testing"
)

func TestConversationService_CreateConversation(t *testing.T) {
	t.Run("return error when BatchWriteItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(errors.New("error"))
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.CreateConversation(test.Model.Account, "sdgsdg", []string{test.ConversationID})

		// Assert
		assert.Error(t, err)
		assert.Empty(t, msg.ID)
	})
	t.Run("returns error if participantIDs contains ID of sender", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.CreateConversation(test.Model.Account, "sdgsdg", []string{test.Model.Account.ID, "efg"})

		// Assert
		assert.Error(t, err)
		assert.Empty(t, msg.ID)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.CreateConversation(test.Model.Account, "sdgsdg", []string{"cbd", "efg"})

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.Equal(t, "sdgsdg", msg.CipherText)
		assert.Equal(t, test.Model.Account.ID, msg.SenderID)
		mockStorage.AssertExpectations(t)
		writeArgs := mockStorage.Calls[0].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, writeArgs, 4) // 1 message + meta for each participant and 1 for sender
	})
}

func TestConversationService_GetConversation(t *testing.T) {
	t.Run("return error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return([]storage.Resource{}, errors.New("error"))
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		conv, err := service.GetConversation(test.Model.Account, test.ConversationID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, conv.Messages)
		assert.Empty(t, conv.Participants)
	})
	t.Run("returns error if not found in storage", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return([]storage.Resource{}, nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		conv, err := service.GetConversation(test.Model.Account, test.ConversationID)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrConversationNotFound)
		assert.Empty(t, conv.Messages)
		assert.Empty(t, conv.Participants)
	})
	t.Run("returns error if Account is not participant", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return(test.Resources, nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		acc := models.Account{ID: "xxx"}
		conv, err := service.GetConversation(acc, test.ConversationID)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		assert.Empty(t, conv.Messages)
		assert.Empty(t, conv.Participants)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := models.ParticipantPrimaryKey(test.ConversationID, "")
		mockStorage.On("QueryItems", primKey.PartitionKey, "", storage.QueryBeginsWith).Return(test.Resources, nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		conv, err := service.GetConversation(test.Model.Account, test.ConversationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, conv.Participants)
		assert.NotEmpty(t, conv.Messages)
		mockStorage.AssertExpectations(t)
	})
}

func TestConversationService_SendMessage(t *testing.T) {
	t.Run("return error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return([]storage.Resource{}, errors.New("error"))
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.SendMessage(test.Model.Account, test.ConversationID, "sdgsdgsdg")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, msg.ID)
	})
	t.Run("returns error if not found in storage", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return([]storage.Resource{}, nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.SendMessage(test.Model.Account, test.ConversationID, "sdgsdgsdg")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrConversationNotFound)
		assert.Empty(t, msg.ID)
	})
	t.Run("returns error if Account is not participant", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return(test.Resources, nil)
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		acc := models.Account{ID: "xxx"}
		msg, err := service.SendMessage(acc, test.ConversationID, "sdgsdgsdg")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		assert.Empty(t, msg.ID)
	})
	t.Run("returns error if BatchWriteItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return(test.Resources, nil)
		mockStorage.On("BatchWriteItems", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
		service := NewConversationService(mockStorage, new(test.MockWebsocketManager))

		// Act
		msg, err := service.SendMessage(test.Model.Account, test.ConversationID, "sdgsdgsdg")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, msg.ID)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		participant2 := storage.Resource{PrimaryKey: models.ParticipantPrimaryKey(test.ConversationID, "879"), Name: test.StringPtr("asdgfadfgasd")}
		participants := []storage.Resource{test.Resource.Participant, participant2}
		mockStorage.On("QueryItems", test.Resource.Participant.PartitionKey, "acc#", storage.QueryBeginsWith).Return(participants, nil)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		mockWebsocket := new(test.MockWebsocketManager)
		mockWebsocket.On("SendToClient", "879", mock.Anything).Return(nil)
		service := NewConversationService(mockStorage, mockWebsocket)

		// Act
		msg, err := service.SendMessage(test.Model.Account, test.ConversationID, "sdgsdgsdg")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "sdgsdgsdg", msg.CipherText)
		assert.Equal(t, test.Model.Account.ID, msg.SenderID)
		mockStorage.AssertExpectations(t)
		writeArgs := mockStorage.Calls[1].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, writeArgs, 3)         // 1 message + 2 participant
		mockWebsocket.AssertExpectations(t) // SendToClient should have been called only for participant2
	})
}
