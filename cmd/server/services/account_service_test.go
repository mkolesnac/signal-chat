package services

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services/mocks"
	"signal-chat/cmd/server/storage"
	"testing"
)

func TestAccountService_CreateAccount(t *testing.T) {
	t.Run("returns error when invalid signed prekey signature", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		service := NewAccountsService(mockStorage)
		signedPreKey := test.signedPreKey1
		signedPreKey.Signature = randomBytes(64) // invalid signature

		// Act
		_, err := service.CreateAccount("test", "test123", test.identityKey, signedPreKey, nil)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSignature)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("returns error BatchWriteItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(errors.New("test error"))
		service := NewAccountsService(mockStorage)

		// Act
		_, err := service.CreateAccount("test", "test123", test.identityKey, test.signedPreKey1, nil)

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when no prekeys in request", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewAccountsService(mockStorage)

		// Act
		acc, err := service.CreateAccount("test", "test123", test.identityKey, test.signedPreKey1, nil)

		// Assert
		assert.NotNil(t, acc)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		args := mockStorage.Calls[0].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, args, 3) // only 3 items were written
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewAccountsService(mockStorage)

		// Act
		acc, err := service.CreateAccount("test", "test123", test.identityKey, test.signedPreKey1, []models.PreKey{test.preKey1})

		// Assert
		assert.NotNil(t, acc)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		args := mockStorage.Calls[0].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, args, 4) // 4 items were written
	})
}

func TestAccountService_GetAccount(t *testing.T) {
	t.Run("returns error when account doesn't exist", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("GetItem", mock.Anything, mock.Anything).Return(storage.ErrNotFound)
		service := NewAccountsService(mockStorage)

		// Act
		_, err := service.GetAccount("123")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotFound)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		res := storage.Resource{
			PrimaryKey:     models.GetAccountPrimaryKey("123"),
			Name:           stringPtr("test User"),
			SignedPreKeyID: stringPtr("abcedf"),
		}
		mockStorage.On("GetItem", res.PartitionKey, res.SortKey).Return(res)
		service := NewAccountsService(mockStorage)

		// Act
		acc, err := service.GetAccount("123")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "123", acc.ID)
		assert.Equal(t, "test User", acc.Name)
		assert.Equal(t, "abcedf", acc.SignedPreKeyID)
		mockStorage.AssertExpectations(t)
	})
}

func TestAccountService_GetSession(t *testing.T) {
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
		service := NewAccountsService(mockStorage)

		// Act
		a, err := service.GetSession(test.account)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, a.Account.ID, "")
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, "", mock.Anything).Return(testResources)
		service := NewAccountsService(mockStorage)

		// Act
		actual, err := service.GetSession(test.account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.account.ID, actual.Account.ID)
		assert.Len(t, actual.Conversations, 2)
		assert.Equal(t, "abc", actual.Conversations[0].ID)
		assert.Equal(t, "edf", actual.Conversations[1].ID)
	})
}

func TestAccountService_GetKeyBundle(t *testing.T) {
	t.Run("returns error when account doesn't exist", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		accKey := models.GetAccountPrimaryKey("123")
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(storage.ErrNotFound)
		service := NewAccountsService(mockStorage)

		// Act
		_, err := service.GetKeyBundle("123")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotFound)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		accKey := models.GetAccountPrimaryKey(test.account.ID)
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(test.account)
		mockStorage.On("QueryItems", accKey.PartitionKey, mock.Anything, mock.Anything).Return(errors.New("test error"))
		service := NewAccountsService(mockStorage)

		// Act
		actual, err := service.GetKeyBundle(test.account.ID)

		// Assert
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrAccountNotFound)
		assert.Equal(t, nil, actual.IdentityKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns error when DeleteItem fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		accKey := models.GetAccountPrimaryKey(test.account.ID)
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(test.account)
		mockStorage.On("QueryItems", accKey.PartitionKey, mock.Anything, mock.Anything).Return(testResources)
		mockStorage.On("DeleteItem", mock.Anything, mock.Anything).Return(errors.New("test error"))
		service := NewAccountsService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.account.ID)

		// Assert
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrAccountNotFound)
		assert.Equal(t, nil, bundle.IdentityKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when multiple signed prekeys in database", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		accKey := models.GetAccountPrimaryKey(test.account.ID)
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(test.account)
		mockStorage.On("QueryItems", mock.Anything, "", mock.Anything).Return(testResources)
		mockStorage.On("DeleteItem", mock.Anything, mock.Anything).Return(nil)
		service := NewAccountsService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.account.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.identityKey, bundle.IdentityKey)
		assert.Equal(t, test.account.SignedPreKeyID, bundle.SignedPreKey.KeyID)
		assert.Equal(t, test.signedPreKey1.PublicKey, bundle.SignedPreKey.PublicKey)
		assert.NotNil(t, bundle.PreKey.PublicKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when no prekeys in database", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		accKey := models.GetAccountPrimaryKey(test.account.ID)
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(test.account)
		res := []storage.Resource{testResources[0], testResources[1], testResources[2], testResources[3]}
		mockStorage.On("QueryItems", mock.Anything, "", mock.Anything).Return(res)
		mockStorage.On("DeleteItem", mock.Anything, mock.Anything).Return(nil)
		service := NewAccountsService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.account.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.identityKey, bundle.IdentityKey)
		assert.Equal(t, test.account.SignedPreKeyID, bundle.SignedPreKey.KeyID)
		assert.Equal(t, test.signedPreKey1.PublicKey, bundle.SignedPreKey.PublicKey)
		assert.Nil(t, bundle.PreKey.PublicKey)
		mockStorage.AssertExpectations(t)
	})
}

func TestAccountService_GetPreKeyCount(t *testing.T) {
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, "", storage.QueryBeginsWith).Return(errors.New("test error"))
		service := NewAccountsService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.account)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns 0 when no prekeys available", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		primKey := models.GetPreKeyPrimaryKey(models.GetAccountPrimaryKey(test.account.ID), "")
		res := []storage.Resource{testResources[1], testResources[2], testResources[3]}
		mockStorage.On("QueryItems", primKey.PartitionKey, "", storage.QueryBeginsWith).Return(res)
		service := NewAccountsService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns count when multiple prekeys available", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		primKey := models.GetPreKeyPrimaryKey(models.GetAccountPrimaryKey(test.account.ID), "")
		res := []storage.Resource{testResources[1], testResources[2], testResources[3], testResources[4], testResources[5]}
		mockStorage.On("QueryItems", primKey.PartitionKey, "", storage.QueryBeginsWith).Return(res)
		service := NewAccountsService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		mockStorage.AssertExpectations(t)
	})
}

func TestAccountService_UploadNewPreKeys(t *testing.T) {
	t.Run("returns error when fails to retrieve identity key", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		primKey := models.GetIdentityKeyPrimaryKey(models.GetAccountPrimaryKey(test.account.ID))
		mockStorage.On("GetItem", primKey.PartitionKey, primKey.SortKey).Return(storage.ErrNotFound)
		service := NewAccountsService(mockStorage)

		// Act
		err := service.UploadNewPreKeys(test.account, test.signedPreKey2, []models.PreKey{test.preKey1, test.preKey2})

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("returns error when invalid signed preKey signature", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		primKey := models.GetIdentityKeyPrimaryKey(models.GetAccountPrimaryKey(test.account.ID))
		mockStorage.On("GetItem", primKey.PartitionKey, primKey.SortKey).Return(test.identityKey)
		service := NewAccountsService(mockStorage)
		signedPreKey := test.signedPreKey2
		signedPreKey.Signature = randomBytes(64)

		// Act
		err := service.UploadNewPreKeys(test.account, signedPreKey, []models.PreKey{test.preKey1, test.preKey2})

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("updates account's signedPreKey ID on success", func(t *testing.T) {
		// Arrange
		mockStorage := new(mocks.MockStorage)
		identityPrimKey := models.GetIdentityKeyPrimaryKey(models.GetAccountPrimaryKey(test.account.ID))
		mockStorage.On("GetItem", identityPrimKey.PartitionKey, identityPrimKey.SortKey).Return(test.identityKey)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		accPrimKey := models.GetAccountPrimaryKey(test.account.ID)
		mockStorage.On("UpdateItem", accPrimKey.PartitionKey, accPrimKey.SortKey, mock.Anything).Return(nil)
		service := NewAccountsService(mockStorage)

		// Act
		err := service.UploadNewPreKeys(test.account, test.signedPreKey2, []models.PreKey{test.preKey1, test.preKey2})

		// Assert
		assert.NoError(t, err)
		batchWriteArgs := mockStorage.Calls[1].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, batchWriteArgs, 3) // 3 resources should have been written to database
		updateArgs := mockStorage.Calls[2].Arguments.Get(2).(map[string]interface{})
		assert.Equal(t, test.signedPreKey2.KeyID, updateArgs["SignedPreKeyID"])
		mockStorage.AssertExpectations(t)
	})
}
