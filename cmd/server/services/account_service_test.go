package services

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/services/mocks"
	"signal-chat/internal/api"
	"testing"
)

func TestAccountService_CreateAccount(t *testing.T) {
	// Define the request object
	req := api.CreateAccountRequest{
		IdentityPublicKey: TestingIdentityKey.PublicKey[:],
		SignedPreKey:      api.SignedPreKeyRequest{KeyID: TestingSignedPreKey.GetID(), PublicKey: TestingSignedPreKey.PublicKey[:], Signature: TestingSignedPreKey.Signature[:]},
	}

	t.Run("successfully creates account", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountService(mockStorage)
		// Arrange
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Account")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.IdentityKey")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(nil)

		// Act
		id, err := service.CreateAccount("Test", "test", req)

		// Assert
		assert.NotEmpty(t, id)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})
	t.Run("error writing account", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountService(mockStorage)
		// Arrange
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Account")).Return(errors.New("write error"))
		mockStorage.On("WriteItem", mock.Anything).Return(nil)

		// Act
		id, err := service.CreateAccount("Test", "test", req)

		// Assert
		assert.Empty(t, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write account")
	})
	t.Run("error writing identity key", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountService(mockStorage)
		// Arrange
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.IdentityKey")).Return(errors.New("write error"))
		mockStorage.On("WriteItem", mock.Anything).Return(nil)

		// Act
		id, err := service.CreateAccount("Test", "test", req)

		// Assert
		assert.Empty(t, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write identity key")
	})
	t.Run("error writing signed pre key", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountService(mockStorage)
		// Arrange
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(errors.New("write error"))
		mockStorage.On("WriteItem", mock.Anything).Return(nil)

		// Act
		id, err := service.CreateAccount("Test", "test", req)

		// Assert
		assert.Empty(t, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write signed pre key")
	})
}
