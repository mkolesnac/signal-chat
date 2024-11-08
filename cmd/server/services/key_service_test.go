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

func TestKeyService_GetPreKeyCount(t *testing.T) {
	t.Run("when prekeys are found", func(t *testing.T) {
		mockStorage := new(mocks.MockStorage)
		keyService := NewKeyService(mockStorage, new(mocks.MockAccountService))
		preKeys := []models.PreKey{{ID: "1"}, {ID: "2"}}
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, storage.BEGINS_WITH, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(3).(*[]models.PreKey) = preKeys
			}).
			Return(nil)

		// Act
		count, err := keyService.GetPreKeyCount("accountID")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
	t.Run("error when no prekeys found", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		keyService := NewKeyService(mockStorage, new(mocks.MockAccountService))
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, storage.BEGINS_WITH, mock.Anything).
			Return(errors.New("query error"))

		// Act
		count, err := keyService.GetPreKeyCount("accountID")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestKeyService_GetPublicKeys(t *testing.T) {
	t.Run("when prekeys are found", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("GetItem", TestingAccount.PartitionKey, TestingIdentityKey.SortKey, mock.AnythingOfType("*models.IdentityKey")).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*models.IdentityKey) = *TestingIdentityKey
			}).Return(nil)
		mockStorage.On("GetItem", TestingAccount.PartitionKey, TestingSignedPreKey.SortKey, mock.AnythingOfType("*models.SignedPreKey")).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*models.SignedPreKey) = *TestingSignedPreKey
			}).Return(nil)
		mockStorage.On("QueryItems", TestingAccount.PartitionKey, models.PreKeySortKey(""), storage.BEGINS_WITH, mock.AnythingOfType("*[]models.PreKey")).
			Run(func(args mock.Arguments) {
				*args.Get(3).(*[]models.PreKey) = []models.PreKey{*TestingPreKey1}
			}).Return(nil)
		mockStorage.On("DeleteItem", TestingAccount.PartitionKey, TestingPreKey1.SortKey).Return(nil)

		// Act
		response, err := keyService.GetPublicKeys(TestingAccount.GetID())

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, TestingIdentityKey.PublicKey, response.IdentityPublicKey)
		assert.Equal(t, TestingSignedPreKey.GetID(), response.SignedPreKey.KeyID)
		assert.Equal(t, TestingSignedPreKey.PublicKey, response.SignedPreKey.PublicKey)
		assert.Equal(t, TestingSignedPreKey.GetID(), response.PreKey.KeyID)
		assert.Equal(t, TestingSignedPreKey.PublicKey, response.PreKey.PublicKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns response without prekeys when no prekeys found", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("GetItem", TestingAccount.PartitionKey, TestingIdentityKey.SortKey, mock.AnythingOfType("*models.IdentityKey")).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*models.IdentityKey) = *TestingIdentityKey
			}).Return(nil)
		mockStorage.On("GetItem", TestingAccount.PartitionKey, TestingSignedPreKey.SortKey, mock.AnythingOfType("*models.SignedPreKey")).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*models.SignedPreKey) = *TestingSignedPreKey
			}).Return(nil)
		mockStorage.On("QueryItems", TestingAccount.PartitionKey, models.PreKeySortKey(""), storage.BEGINS_WITH, mock.AnythingOfType("*[]models.PreKey")).Return(nil)

		// Act
		response, err := keyService.GetPublicKeys(TestingAccount.GetID())

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Nil(t, response.PreKey)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "DeleteItem")
	})
	t.Run("returns argument error when no account found", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		mockAccounts.On("GetAccount", mock.Anything).Return(nil, ErrAccountNotFound)

		// Act
		response, err := keyService.GetPublicKeys(TestingAccount.GetID())

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrAccountNotFound)
		mockStorage.AssertNotCalled(t, "DeleteItem")
	})
}

func TestKeyService_VerifySignature(t *testing.T) {
	t.Run("valid signature", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		keyService := NewKeyService(mockStorage, new(mocks.MockAccountService))
		// Mock the identity key
		mockStorage.On("GetItem", TestingIdentityKey.PartitionKey, TestingIdentityKey.SortKey, mock.AnythingOfType("*models.IdentityKey")).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*models.IdentityKey) = *TestingIdentityKey
			}).
			Return(nil)

		// Act
		result, err := keyService.VerifySignature(TestingAccount.GetID(), TestingSignedPreKey.PublicKey[:], TestingSignedPreKey.Signature[:])

		// Assert
		assert.NoError(t, err)
		assert.True(t, result)
	})
	t.Run("when identity key not found", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		keyService := NewKeyService(mockStorage, new(mocks.MockAccountService))
		// Mock error in GetItem
		mockStorage.On("GetItem", mock.Anything, mock.Anything, mock.AnythingOfType("*models.IdentityKey")).
			Return(errors.New("item not found"))

		// Act
		result, err := keyService.VerifySignature(TestingAccount.GetID(), TestingSignedPreKey.PublicKey[:], TestingSignedPreKey.Signature[:])

		// Asser
		assert.Error(t, err)
		assert.False(t, result)
	})
}

func TestKeyService_UploadNewPreKeys(t *testing.T) {
	// Define the request object
	req := api.UploadPreKeysRequest{
		SignedPreKey: api.SignedPreKeyRequest{KeyID: TestingSignedPreKey.GetID(), PublicKey: TestingSignedPreKey.PublicKey[:], Signature: TestingSignedPreKey.Signature[:]},
		PreKeys: []api.PreKeyRequest{
			{KeyID: TestingPreKey1.ID, PublicKey: TestingPreKey1.PublicKey[:]},
			{KeyID: TestingPreKey2.ID, PublicKey: TestingPreKey2.PublicKey[:]},
		},
	}

	t.Run("successful upload of prekeys", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		// Define the custom matcher for *models.Account with SignedPreKeyID == TestingSignedPreKey.ID
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		accountMatcher := mock.MatchedBy(func(item storage.PrimaryKeyProvider) bool {
			account, ok := item.(*models.Account)
			return ok && account.SignedPreKeyID == TestingSignedPreKey.GetID()
		})
		mockStorage.On("WriteItem", accountMatcher).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(nil)
		mockStorage.On("BatchWriteItems", mock.AnythingOfType("[]storage.PrimaryKeyProvider")).Return(nil)

		// Act
		err := keyService.UploadNewPreKeys("123", req)

		// Assert
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})
	t.Run("error writing signed prekey", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Account")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(errors.New("write error"))
		mockStorage.On("BatchWriteItems", mock.AnythingOfType("[]storage.PrimaryKeyProvider")).Return(nil)

		// Act
		err := keyService.UploadNewPreKeys("123", req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write signed pre key")
	})
	t.Run("error writing prekeys", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockAccounts := new(mocks.MockAccountService)
		keyService := NewKeyService(mockStorage, mockAccounts)
		mockAccounts.On("GetAccount", TestingAccount.GetID()).Return(TestingAccount, nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.Account")).Return(nil)
		mockStorage.On("WriteItem", mock.AnythingOfType("*models.SignedPreKey")).Return(nil)
		mockStorage.On("BatchWriteItems", mock.AnythingOfType("[]storage.PrimaryKeyProvider")).Return(errors.New("write error"))

		// Act
		err := keyService.UploadNewPreKeys("123", req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to batch write pre keys")
	})
}
