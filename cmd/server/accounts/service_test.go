package accounts

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/storage/mocks"
	"testing"
)

func TestAccountService_CreateAccount(t *testing.T) {
	// Define the request object
	req := CreateAccountRequest{
		IdentityPublicKey: TestingIdentityKey.PublicKey[:],
		SignedPreKey:      SignedPreKeyRequest{KeyID: TestingSignedPreKey.ID, PublicKey: TestingSignedPreKey.PublicKey[:], Signature: TestingSignedPreKey.Signature[:]},
	}

	t.Run("successfuly creates account", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountService(mockStorage)
		// Arrange
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Account")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.IdentityKey")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(nil)

		// Act
		err := service.CreateAccount("123", "test", req)

		// Assert
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
		err := service.CreateAccount("123", "test", req)

		// Assert
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
		err := service.CreateAccount("123", "test", req)

		// Assert
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
		err := service.CreateAccount("123", "test", req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write signed pre key")
	})
}
