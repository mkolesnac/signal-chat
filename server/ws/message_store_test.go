package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"signal-chat/internal/api"
)

func TestMessageStore_Store(t *testing.T) {
	t.Run("should store messages successfully", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		messages := []*api.WSMessage{
			{
				ID:   "msg1",
				Type: api.MessageTypeNewMessage,
				Data: json.RawMessage(`{"content": "Hello"}`),
			},
			{
				ID:   "msg2",
				Type: api.MessageTypeNewConversation,
				Data: json.RawMessage(`{"conversationId": "conv1"}`),
			},
		}

		// Act
		err := store.Store(messages)

		// Assert
		require.NoError(t, err)

		// Verify messages were stored
		storedMessages, err := store.LoadAll()
		require.NoError(t, err)
		require.Len(t, storedMessages, 2)

		assert.Equal(t, "msg1", storedMessages[0].ID)
		assert.Equal(t, api.MessageTypeNewMessage, storedMessages[0].Type)
		assert.JSONEq(t, `{"content": "Hello"}`, string(storedMessages[0].Data))

		assert.Equal(t, "msg2", storedMessages[1].ID)
		assert.Equal(t, api.MessageTypeNewConversation, storedMessages[1].Type)
		assert.JSONEq(t, `{"conversationId": "conv1"}`, string(storedMessages[1].Data))
	})

	t.Run("should handle empty message list", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Act
		err := store.Store([]*api.WSMessage{})

		// Assert
		require.NoError(t, err)

		storedMessages, err := store.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, storedMessages)
	})
}

func TestMessageStore_Delete(t *testing.T) {
	t.Run("should delete specified messages", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Store some messages first
		messages := []*api.WSMessage{
			{ID: "msg1", Type: api.MessageTypeNewMessage, Data: json.RawMessage(`{"content": "Hello"}`)},
			{ID: "msg2", Type: api.MessageTypeNewMessage, Data: json.RawMessage(`{"content": "World"}`)},
			{ID: "msg3", Type: api.MessageTypeNewMessage, Data: json.RawMessage(`{"content": "!"}`)},
		}
		err := store.Store(messages)
		require.NoError(t, err)

		// Act
		err = store.Delete([]string{"msg1", "msg3"})

		// Assert
		require.NoError(t, err)

		storedMessages, err := store.LoadAll()
		require.NoError(t, err)
		require.Len(t, storedMessages, 1)
		assert.Equal(t, "msg2", storedMessages[0].ID)
	})

	t.Run("should handle empty message IDs list", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Act
		err := store.Delete([]string{})

		// Assert
		require.NoError(t, err)
	})

	t.Run("should handle non-existent message IDs", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Act
		err := store.Delete([]string{"non-existent-1", "non-existent-2"})

		// Assert
		require.NoError(t, err)
	})
}

func TestMessageStore_LoadAll(t *testing.T) {
	t.Run("should load all messages", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Store messages with different types
		messages := []*api.WSMessage{
			{ID: "msg1", Type: api.MessageTypeNewMessage, Data: json.RawMessage(`{"content": "Hello"}`)},
			{ID: "msg2", Type: api.MessageTypeNewConversation, Data: json.RawMessage(`{"conversationId": "conv1"}`)},
			{ID: "msg3", Type: api.MessageTypeParticipantAdded, Data: json.RawMessage(`{"participantId": "user1"}`)},
		}
		err := store.Store(messages)
		require.NoError(t, err)

		// Act
		loadedMessages, err := store.LoadAll()

		// Assert
		require.NoError(t, err)
		require.Len(t, loadedMessages, 3)

		// Verify each message
		assert.Equal(t, "msg1", loadedMessages[0].ID)
		assert.Equal(t, api.MessageTypeNewMessage, loadedMessages[0].Type)
		assert.JSONEq(t, `{"content": "Hello"}`, string(loadedMessages[0].Data))

		assert.Equal(t, "msg2", loadedMessages[1].ID)
		assert.Equal(t, api.MessageTypeNewConversation, loadedMessages[1].Type)
		assert.JSONEq(t, `{"conversationId": "conv1"}`, string(loadedMessages[1].Data))

		assert.Equal(t, "msg3", loadedMessages[2].ID)
		assert.Equal(t, api.MessageTypeParticipantAdded, loadedMessages[2].Type)
		assert.JSONEq(t, `{"participantId": "user1"}`, string(loadedMessages[2].Data))
	})

	t.Run("should return empty slice when no messages exist", func(t *testing.T) {
		// Arrange
		db, cleanup := testDB(t)
		defer cleanup()

		store := &MessageStore{
			db:       db,
			clientID: "test-client",
		}

		// Act
		messages, err := store.LoadAll()

		// Assert
		require.NoError(t, err)
		assert.Empty(t, messages)
	})
}
