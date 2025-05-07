package encryption

import (
	"errors"
	"signal-chat/client/api"
	"signal-chat/client/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_InitializeKeyStore(t *testing.T) {
	t.Run("should initialize key store with valid keys", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		err := db.Open("test-user")
		require.NoError(t, err)

		apiClient := api.NewStubClient()
		manager := NewEncryptionManager(db, apiClient)

		// Act
		keyBundle, err := manager.InitializeKeyStore()

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, keyBundle.IdentityKey)
		assert.NotEmpty(t, keyBundle.SignedPreKey.PublicKey)
		assert.NotEmpty(t, keyBundle.SignedPreKey.Signature)
		assert.NotEmpty(t, keyBundle.PreKeys)
		assert.Greater(t, len(keyBundle.PreKeys), 0)
	})

	t.Run("should return error when database fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.WriteErr = errors.New("database error")
		apiClient := api.NewStubClient()
		manager := NewEncryptionManager(db, apiClient)

		// Act
		_, err := manager.InitializeKeyStore()

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to")
	})
}

func TestManager_CreateEncryptionGroup(t *testing.T) {
	t.Run("should create encryption group with key distribution messages for each participant", func(t *testing.T) {
		// Arrange
		apiClient := api.NewFakeClient()

		// Set up two users: sender and receiver, each with its own database and encryption manager
		user1DB := database.NewFake()
		err := user1DB.Open("sender-user")
		require.NoError(t, err)
		user1Manager := NewEncryptionManager(user1DB, apiClient)
		user1Bundle, err := user1Manager.InitializeKeyStore()
		require.NoError(t, err)
		user1, err := apiClient.SignUp("user1", "password", user1Bundle)
		require.NoError(t, err)

		user2DB := database.NewFake()
		err = user2DB.Open("receiver-user")
		require.NoError(t, err)
		user2Manager := NewEncryptionManager(user2DB, apiClient)
		user2Bundle, err := user2Manager.InitializeKeyStore()
		require.NoError(t, err)
		user2, err := apiClient.SignUp("user2", "password", user2Bundle)
		require.NoError(t, err)

		// Set up the main user who will be creating the encryption group
		mainUserDB := database.NewFake()
		err = mainUserDB.Open("main-user")
		require.NoError(t, err)
		mainUserManager := NewEncryptionManager(mainUserDB, apiClient)
		mainUserBundle, err := mainUserManager.InitializeKeyStore()
		require.NoError(t, err)
		_, err = apiClient.SignUp("main-user", "password", mainUserBundle)
		require.NoError(t, err)

		// Act
		keyMessages, err := mainUserManager.CreateEncryptionGroup("group1", []string{user1.UserID, user2.UserID})

		// Assert
		require.NoError(t, err)
		require.Len(t, keyMessages, 2)
		assert.Contains(t, keyMessages, user1.UserID)
		assert.Contains(t, keyMessages, user2.UserID)
		assert.NotEmpty(t, keyMessages[user1.UserID])
		assert.NotEmpty(t, keyMessages[user2.UserID])
	})

	t.Run("should return error when API client fails", func(t *testing.T) {
		// Arrange
		apiClient := api.NewStubClient()
		apiClient.GetPreKeyBundleError = errors.New("API error")

		db := database.NewFake()
		err := db.Open("test-user")
		require.NoError(t, err)
		manager := NewEncryptionManager(db, apiClient)
		_, err = manager.InitializeKeyStore()
		require.NoError(t, err)

		// Act
		groupID := "group1"
		recipientIDs := []string{"user1"}
		_, err = manager.CreateEncryptionGroup(groupID, recipientIDs)

		// Assert
		require.Error(t, err)
	})
}

func TestManager_ProcessSenderKeyDistributionMessage(t *testing.T) {
	t.Run("should return error for corrupted message", func(t *testing.T) {
		// Arrange
		receiverDB := database.NewFake()
		err := receiverDB.Open("receiver-user")
		require.NoError(t, err)
		apiClient := api.NewFakeClient()

		receiverManager := NewEncryptionManager(receiverDB, apiClient)
		_, err = receiverManager.InitializeKeyStore()
		require.NoError(t, err)

		// Act
		corruptedMessage := []byte("corrupted-message")
		err = receiverManager.ProcessSenderKeyDistributionMessage("group1", "sender", corruptedMessage)

		// Assert
		require.Error(t, err)
	})
}

func TestManager_GroupEncryptDecrypt(t *testing.T) {
	t.Run("should encrypt and decrypt messages in a group", func(t *testing.T) {
		// Arrange
		apiClient := api.NewFakeClient()

		// Set up two users: sender and receiver, each with its own database and encryption manager
		senderDB := database.NewFake()
		err := senderDB.Open("sender-user")
		require.NoError(t, err)
		senderManager := NewEncryptionManager(senderDB, apiClient)
		senderBundle, err := senderManager.InitializeKeyStore()
		require.NoError(t, err)
		sender, err := apiClient.SignUp("sender", "password", senderBundle)
		require.NoError(t, err)

		receiverDB := database.NewFake()
		err = receiverDB.Open("receiver-user")
		require.NoError(t, err)
		receiverManager := NewEncryptionManager(receiverDB, apiClient)
		receiverBundle, err := receiverManager.InitializeKeyStore()
		require.NoError(t, err)
		receiver, err := apiClient.SignUp("receiver", "password", receiverBundle)
		require.NoError(t, err)

		// Create an encryption group
		groupID := "group1"
		keyMessages, err := senderManager.CreateEncryptionGroup(groupID, []string{receiver.UserID})
		require.NoError(t, err)
		require.Contains(t, keyMessages, receiver.UserID)

		err = receiverManager.ProcessSenderKeyDistributionMessage(groupID, sender.UserID, keyMessages[receiver.UserID])
		require.NoError(t, err)

		// Act
		plaintext := []byte("hello secure world")
		encryptedMsg, err := senderManager.GroupEncrypt(groupID, plaintext)
		require.NoError(t, err)

		decryptedMsg, err := receiverManager.GroupDecrypt(groupID, sender.UserID, encryptedMsg.Serialized)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, plaintext, decryptedMsg.Plaintext)
	})

	t.Run("should fail to decrypt with wrong sender ID", func(t *testing.T) {
		// Arrange
		apiClient := api.NewFakeClient()

		// Set up two users: sender and receiver, each with its own database and encryption manager
		senderDB := database.NewFake()
		err := senderDB.Open("sender-user")
		require.NoError(t, err)
		senderManager := NewEncryptionManager(senderDB, apiClient)
		senderBundle, err := senderManager.InitializeKeyStore()
		require.NoError(t, err)
		sender, err := apiClient.SignUp("sender", "password", senderBundle)
		require.NoError(t, err)

		receiverDB := database.NewFake()
		err = receiverDB.Open("receiver-user")
		require.NoError(t, err)
		receiverManager := NewEncryptionManager(receiverDB, apiClient)
		receiverBundle, err := receiverManager.InitializeKeyStore()
		require.NoError(t, err)
		receiver, err := apiClient.SignUp("receiver", "password", receiverBundle)
		require.NoError(t, err)

		// Create an encryption group
		groupID := "group1"
		keyMessages, err := senderManager.CreateEncryptionGroup(groupID, []string{receiver.UserID})
		require.NoError(t, err)
		require.Contains(t, keyMessages, receiver.UserID)

		err = receiverManager.ProcessSenderKeyDistributionMessage(groupID, sender.UserID, keyMessages[receiver.UserID])
		require.NoError(t, err)

		// Act
		plaintext := []byte("hello secure world")
		encryptedMsg, err := senderManager.GroupEncrypt(groupID, plaintext)
		require.NoError(t, err)

		// Try to decrypt with wrong sender ID
		_, err = receiverManager.GroupDecrypt(groupID, "wrong-sender", encryptedMsg.Serialized)

		// Assert
		require.Error(t, err)
	})

	t.Run("should fail to encrypt message when sender key not found", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		err := db.Open("test-user")
		require.NoError(t, err)
		apiClient := api.NewStubClient()
		manager := NewEncryptionManager(db, apiClient)
		_, err = manager.InitializeKeyStore()
		require.NoError(t, err)

		// Act - try to encrypt without creating group
		_, err = manager.GroupEncrypt("non-existent-group", []byte("test-message"))

		// Assert
		assert.Error(t, err)
	})

	t.Run("should fail to decrypt corrupted message", func(t *testing.T) {
		// Arrange
		apiClient := api.NewFakeClient()

		// Set up two users: sender and receiver, each with its own database and encryption manager
		senderDB := database.NewFake()
		err := senderDB.Open("sender-user")
		require.NoError(t, err)
		senderManager := NewEncryptionManager(senderDB, apiClient)
		senderBundle, err := senderManager.InitializeKeyStore()
		require.NoError(t, err)
		sender, err := apiClient.SignUp("sender", "password", senderBundle)
		require.NoError(t, err)

		receiverDB := database.NewFake()
		err = receiverDB.Open("receiver-user")
		require.NoError(t, err)
		receiverManager := NewEncryptionManager(receiverDB, apiClient)
		receiverBundle, err := receiverManager.InitializeKeyStore()
		require.NoError(t, err)
		receiver, err := apiClient.SignUp("receiver", "password", receiverBundle)
		require.NoError(t, err)

		// Create an encryption group
		groupID := "group1"
		keyMessages, err := senderManager.CreateEncryptionGroup(groupID, []string{receiver.UserID})
		require.NoError(t, err)
		require.Contains(t, keyMessages, receiver.UserID)

		err = receiverManager.ProcessSenderKeyDistributionMessage(groupID, sender.UserID, keyMessages[receiver.UserID])
		require.NoError(t, err)

		// Act - try to decrypt corrupted message
		corruptedCiphertext := []byte("corrupted-ciphertext")
		_, err = receiverManager.GroupDecrypt(groupID, sender.UserID, corruptedCiphertext)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to deserialize")
	})
}
