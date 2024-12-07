package services

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services/test"
	"signal-chat/cmd/server/storage"
	"signal-chat/cmd/server/utils"
	"testing"
)

func TestAccountService_CreateAccount(t *testing.T) {
	t.Run("returns error when invalid signed prekey signature", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		service := NewAccountService(mockStorage)
		signedPreKey := test.Model.SignedPreKey1
		signedPreKey.Signature = utils.RandomBytes(64) // invalid signature

		// Act
		_, err := service.CreateAccount("model", "test123", test.Model.IdentityKey, signedPreKey, nil)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidSignature)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("returns error BatchWriteItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(errors.New("model error"))
		service := NewAccountService(mockStorage)

		// Act
		_, err := service.CreateAccount("model", "test123", test.Model.IdentityKey, test.Model.SignedPreKey1, nil)

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when no prekeys in request", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewAccountService(mockStorage)

		// Act
		acc, err := service.CreateAccount("model", "test123", test.Model.IdentityKey, test.Model.SignedPreKey1, nil)

		// Assert
		assert.NotNil(t, acc)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		args := mockStorage.Calls[0].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, args, 3) // only 3 items were written
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		service := NewAccountService(mockStorage)

		// Act
		acc, err := service.CreateAccount("model", "test123", test.Model.IdentityKey, test.Model.SignedPreKey1, []models.PreKey{test.Model.PreKey1})

		// Assert
		assert.NotNil(t, acc)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		args := mockStorage.Calls[0].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, args, 4) // 4 items were written
	})
}

func TestAccountService_GetAccount(t *testing.T) {
	t.Run("returns error when Account doesn't exist", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("GetItem", mock.Anything, mock.Anything).Return(storage.Resource{}, storage.ErrNotFound)
		service := NewAccountService(mockStorage)

		// Act
		_, err := service.GetAccount("xxx")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotFound)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := test.Resource.Account.PrimaryKey
		mockStorage.On("GetItem", primKey.PartitionKey, primKey.SortKey).Return(test.Resource.Account, nil)
		service := NewAccountService(mockStorage)

		// Act
		acc, err := service.GetAccount(test.Model.Account.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.Model.Account.ID, acc.ID)
		assert.Equal(t, test.Model.Account.Name, acc.Name)
		assert.Equal(t, test.Model.Account.SignedPreKeyID, acc.SignedPreKeyID)
		assert.EqualValues(t, test.Model.Account.PasswordHash, acc.PasswordHash)
		mockStorage.AssertExpectations(t)
	})
}

func TestAccountService_GetSession(t *testing.T) {
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, mock.Anything, mock.Anything).Return([]storage.Resource{}, errors.New("model error"))
		service := NewAccountService(mockStorage)

		// Act
		a, err := service.GetSession(test.Model.Account)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, a.Account.ID, "")
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, "", mock.Anything).Return(test.Resources, nil)
		service := NewAccountService(mockStorage)

		// Act
		actual, err := service.GetSession(test.Model.Account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.Model.Account.ID, actual.Account.ID)
		assert.Len(t, actual.Conversations, 1)
		assert.Equal(t, "123", actual.Conversations[0].ID)
	})
}

func TestAccountService_GetKeyBundle(t *testing.T) {
	t.Run("returns error when Account doesn't exist", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		accKey := models.AccountPrimaryKey("123")
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(storage.Resource{}, storage.ErrNotFound)
		service := NewAccountService(mockStorage)

		// Act
		_, err := service.GetKeyBundle("123")

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAccountNotFound)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		accPrimKey := test.Resource.Account.PrimaryKey
		mockStorage.On("GetItem", accPrimKey.PartitionKey, accPrimKey.SortKey).Return(test.Resource.Account, nil)
		mockStorage.On("QueryItems", test.Resource.PreKey1.PartitionKey, "", storage.QueryBeginsWith).Return([]storage.Resource{}, errors.New("model error"))
		service := NewAccountService(mockStorage)

		// Act
		actual, err := service.GetKeyBundle(test.Model.Account.ID)

		// Assert
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrAccountNotFound)
		assert.Empty(t, actual.IdentityKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns error when DeleteItem fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		accPrimKey := test.Resource.Account.PrimaryKey
		mockStorage.On("GetItem", accPrimKey.PartitionKey, accPrimKey.SortKey).Return(test.Resource.Account, nil)
		mockStorage.On("QueryItems", test.Resource.PreKey1.PartitionKey, "", storage.QueryBeginsWith).Return(test.Resources, nil)
		mockStorage.On("DeleteItem", mock.Anything, mock.Anything).Return(errors.New("error"))
		service := NewAccountService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.Model.Account.ID)

		// Assert
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrAccountNotFound)
		assert.Empty(t, bundle.IdentityKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when multiple signed prekeys in database", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		accPrimKey := test.Resource.Account.PrimaryKey
		mockStorage.On("GetItem", accPrimKey.PartitionKey, accPrimKey.SortKey).Return(test.Resource.Account, nil)
		mockStorage.On("QueryItems", test.Resource.PreKey1.PartitionKey, "", storage.QueryBeginsWith).Return(test.Resources, nil)
		mockStorage.On("DeleteItem", mock.Anything, mock.Anything).Return(nil)
		service := NewAccountService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.Model.Account.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.Model.IdentityKey[:], bundle.IdentityKey)
		assert.Equal(t, test.Model.Account.SignedPreKeyID, bundle.SignedPreKey.KeyID)
		assert.EqualValues(t, test.Model.SignedPreKey1.PublicKey, bundle.SignedPreKey.PublicKey)
		assert.NotNil(t, bundle.PreKey.PublicKey)
		mockStorage.AssertExpectations(t)
	})
	t.Run("test success when no prekeys in database", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		accKey := models.AccountPrimaryKey(test.Model.Account.ID)
		mockStorage.On("GetItem", accKey.PartitionKey, accKey.SortKey).Return(test.Resource.Account, nil)
		res := []storage.Resource{test.Resource.Account, test.Resource.IdentityKey, test.Resource.SignedPreKey1, test.Resource.SignedPreKey2}
		mockStorage.On("QueryItems", test.Resource.PreKey1.PartitionKey, "", storage.QueryBeginsWith).Return(res, nil)
		service := NewAccountService(mockStorage)

		// Act
		bundle, err := service.GetKeyBundle(test.Model.Account.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, test.Model.IdentityKey[:], bundle.IdentityKey)
		assert.Equal(t, test.Model.Account.SignedPreKeyID, bundle.SignedPreKey.KeyID)
		assert.Equal(t, test.Model.SignedPreKey1.PublicKey, bundle.SignedPreKey.PublicKey)
		assert.Nil(t, bundle.PreKey.PublicKey)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "DeleteItem")
	})
}

func TestAccountService_GetPreKeyCount(t *testing.T) {
	t.Run("returns error when QueryItems fails", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		mockStorage.On("QueryItems", mock.Anything, "", storage.QueryBeginsWith).Return([]storage.Resource{}, errors.New("model error"))
		service := NewAccountService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.Model.Account)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns 0 when no prekeys available", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := models.PreKeyPrimaryKey(test.Model.Account.ID, "")
		res := []storage.Resource{test.Resource.IdentityKey, test.Resource.SignedPreKey1, test.Resource.SignedPreKey2}
		mockStorage.On("QueryItems", primKey.PartitionKey, "", storage.QueryBeginsWith).Return(res, nil)
		service := NewAccountService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.Model.Account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		mockStorage.AssertExpectations(t)
	})
	t.Run("returns count when multiple prekeys available", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := models.PreKeyPrimaryKey(test.Model.Account.ID, "")
		mockStorage.On("QueryItems", primKey.PartitionKey, "", storage.QueryBeginsWith).Return(test.Resources, nil)
		service := NewAccountService(mockStorage)

		// Act
		count, err := service.GetPreKeyCount(test.Model.Account)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		mockStorage.AssertExpectations(t)
	})
}

func TestAccountService_UploadNewPreKeys(t *testing.T) {
	t.Run("returns error when fails to retrieve identity key", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := models.IdentityKeyPrimaryKey(test.Model.Account.ID)
		mockStorage.On("GetItem", primKey.PartitionKey, primKey.SortKey).Return(storage.Resource{}, storage.ErrNotFound)
		service := NewAccountService(mockStorage)

		// Act
		err := service.UploadNewPreKeys(test.Model.Account, test.Model.SignedPreKey2, []models.PreKey{test.Model.PreKey1, test.Model.PreKey2})

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("returns error when invalid signed preKey signature", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		primKey := models.IdentityKeyPrimaryKey(test.Model.Account.ID)
		mockStorage.On("GetItem", primKey.PartitionKey, primKey.SortKey).Return(test.Resource.IdentityKey, nil)
		service := NewAccountService(mockStorage)
		signedPreKey := test.Model.SignedPreKey2
		signedPreKey.Signature = utils.RandomBytes(64)

		// Act
		err := service.UploadNewPreKeys(test.Model.Account, signedPreKey, []models.PreKey{test.Model.PreKey1, test.Model.PreKey2})

		// Assert
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "BatchWriteItems", mock.Anything)
	})
	t.Run("updates Account's signedPreKey ID on success", func(t *testing.T) {
		// Arrange
		mockStorage := new(test.MockStorage)
		identityPrimKey := models.IdentityKeyPrimaryKey(test.Model.Account.ID)
		mockStorage.On("GetItem", identityPrimKey.PartitionKey, identityPrimKey.SortKey).Return(test.Resource.IdentityKey, nil)
		mockStorage.On("BatchWriteItems", mock.Anything).Return(nil)
		accPrimKey := models.AccountPrimaryKey(test.Model.Account.ID)
		mockStorage.On("UpdateItem", accPrimKey.PartitionKey, accPrimKey.SortKey, mock.Anything).Return(nil)
		service := NewAccountService(mockStorage)

		// Act
		err := service.UploadNewPreKeys(test.Model.Account, test.Model.SignedPreKey2, []models.PreKey{test.Model.PreKey1, test.Model.PreKey2})

		// Assert
		assert.NoError(t, err)
		batchWriteArgs := mockStorage.Calls[1].Arguments.Get(0).([]storage.Resource)
		assert.Len(t, batchWriteArgs, 3) // 3 resources should have been written to database
		updateArgs := mockStorage.Calls[2].Arguments.Get(2).(map[string]interface{})
		assert.Equal(t, test.Model.SignedPreKey2.KeyID, updateArgs["SignedPreKeyID"])
		mockStorage.AssertExpectations(t)
	})
}
