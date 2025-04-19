package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"signal-chat/internal/api"
	"signal-chat/server/conversation"
)

// TestManager_BroadcastNewMessage tests the BroadcastNewMessage method
func TestManager_BroadcastNewMessage(t *testing.T) {
	t.Run("should broadcast message to all participants except sender", func(t *testing.T) {
		// Arrange
		db, dbClose := testDB(t)
		defer dbClose()

		convRepo := NewMockConversationRepository()
		manager := NewManager(db, convRepo)

		// Create a test conversation
		conv := &conversation.Conversation{
			ParticipantIDs: []string{"user-1", "user-2", "user-3"},
		}
		convRepo.AddConversation("conv-123", conv)

		senderID := "user-1"
		messageID := "msg-123"
		req := api.NewMessageRequest{
			ConversationID:   "conv-123",
			EncryptedMessage: []byte("encrypted-message"),
		}

		// Create fake clients for recipients
		fakeConn1 := NewFakeWebSocketConn()
		fakeConn2 := NewFakeWebSocketConn()
		err := manager.RegisterClient("user-2", fakeConn1)
		require.NoError(t, err)
		err = manager.RegisterClient("user-3", fakeConn2)
		require.NoError(t, err)

		// Act
		err = manager.BroadcastNewMessage(senderID, messageID, req)
		require.NoError(t, err)

		// Wait for messages to be sent
		time.Sleep(100 * time.Millisecond)

		// Assert
		// Check that user-2 received the message
		var msg1 api.WSMessage
		select {
		case msgBytes := <-fakeConn1.writeChan:
			err := json.Unmarshal(msgBytes, &msg1)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent to user-2")
		}

		// Check that user-3 received the message
		var msg2 api.WSMessage
		select {
		case msgBytes := <-fakeConn2.writeChan:
			err := json.Unmarshal(msgBytes, &msg2)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent to user-3")
		}

		// Verify message content
		assert.Equal(t, api.MessageTypeNewMessage, msg1.Type)
		assert.Equal(t, api.MessageTypeNewMessage, msg2.Type)

		var payload1 api.WSNewMessagePayload
		err = json.Unmarshal(msg1.Data, &payload1)
		require.NoError(t, err)

		var payload2 api.WSNewMessagePayload
		err = json.Unmarshal(msg2.Data, &payload2)
		require.NoError(t, err)

		assert.Equal(t, messageID, payload1.MessageID)
		assert.Equal(t, messageID, payload2.MessageID)
		assert.Equal(t, senderID, payload1.SenderID)
		assert.Equal(t, senderID, payload2.SenderID)
		assert.Equal(t, req.ConversationID, payload1.ConversationID)
		assert.Equal(t, req.ConversationID, payload2.ConversationID)
		assert.Equal(t, req.EncryptedMessage, payload1.EncryptedMessage)
		assert.Equal(t, req.EncryptedMessage, payload2.EncryptedMessage)
	})

	t.Run("should store message for offline recipients", func(t *testing.T) {
		// Arrange
		db, dbClose := testDB(t)
		defer dbClose()

		convRepo := NewMockConversationRepository()
		manager := NewManager(db, convRepo)

		// Create a test conversation
		conv := &conversation.Conversation{
			ParticipantIDs: []string{"user-1", "user-2", "user-3"},
		}
		convRepo.AddConversation("conv-123", conv)

		senderID := "user-1"
		messageID := "msg-456"
		req := api.NewMessageRequest{
			ConversationID:   "conv-123",
			EncryptedMessage: []byte("encrypted-message"),
		}

		// Act
		err := manager.BroadcastNewMessage(senderID, messageID, req)
		require.NoError(t, err)

		// Assert
		// Check that messages were stored for offline recipients
		messageStore1 := &MessageStore{
			db:       db,
			clientID: "user-2",
		}
		messages1, err := messageStore1.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages1, 1)
		assert.Equal(t, api.MessageTypeNewMessage, messages1[0].Type)

		var payload1 api.WSNewMessagePayload
		err = json.Unmarshal(messages1[0].Data, &payload1)
		require.NoError(t, err)
		assert.Equal(t, messageID, payload1.MessageID)
		assert.Equal(t, senderID, payload1.SenderID)
		assert.Equal(t, req.ConversationID, payload1.ConversationID)
		assert.Equal(t, req.EncryptedMessage, payload1.EncryptedMessage)

		messageStore2 := &MessageStore{
			db:       db,
			clientID: "user-3",
		}
		messages2, err := messageStore2.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages2, 1)
		assert.Equal(t, api.MessageTypeNewMessage, messages2[0].Type)

		var payload2 api.WSNewMessagePayload
		err = json.Unmarshal(messages2[0].Data, &payload2)
		require.NoError(t, err)
		assert.Equal(t, messageID, payload2.MessageID)
		assert.Equal(t, senderID, payload2.SenderID)
		assert.Equal(t, req.ConversationID, payload2.ConversationID)
		assert.Equal(t, req.EncryptedMessage, payload2.EncryptedMessage)
	})

	t.Run("should return error for non-existent conversation", func(t *testing.T) {
		// Arrange
		db, dbClose := testDB(t)
		defer dbClose()

		convRepo := NewMockConversationRepository()
		manager := NewManager(db, convRepo)

		senderID := "user-1"
		messageID := "msg-789"
		req := api.NewMessageRequest{
			ConversationID:   "non-existent-conv",
			EncryptedMessage: []byte("encrypted-message"),
		}

		// Act
		err := manager.BroadcastNewMessage(senderID, messageID, req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get conv")
	})
}

// TestManager_BroadcastNewConversation tests the BroadcastNewConversation method
func TestManager_BroadcastNewConversation(t *testing.T) {
	t.Run("should broadcast new conversation to all recipients", func(t *testing.T) {
		// Arrange
		db, dbClose := testDB(t)
		defer dbClose()

		convRepo := NewMockConversationRepository()
		manager := NewManager(db, convRepo)

		senderID := "user-1"
		req := api.NewConversationRequest{
			ConversationID: "conv-123",
			OtherParticipants: []api.Participant{
				{
					ID:                     "user-2",
					KeyDistributionMessage: []byte("key-distribution-message"),
				},
				{
					ID:                     "user-3",
					KeyDistributionMessage: []byte("key-distribution-message"),
				},
			},
		}

		// Create fake clients for recipients
		fakeConn1 := NewFakeWebSocketConn()
		fakeConn2 := NewFakeWebSocketConn()
		err := manager.RegisterClient("user-2", fakeConn1)
		require.NoError(t, err)
		err = manager.RegisterClient("user-3", fakeConn2)
		require.NoError(t, err)

		// Act
		err = manager.BroadcastNewConversation(senderID, req)
		require.NoError(t, err)

		// Wait for messages to be sent
		time.Sleep(100 * time.Millisecond)

		// Assert
		// Check that user-2 received the message
		var msg1 api.WSMessage
		select {
		case msgBytes := <-fakeConn1.writeChan:
			err := json.Unmarshal(msgBytes, &msg1)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent to user-2")
		}

		// Check that user-3 received the message
		var msg2 api.WSMessage
		select {
		case msgBytes := <-fakeConn2.writeChan:
			err := json.Unmarshal(msgBytes, &msg2)
			require.NoError(t, err)
		default:
			t.Fatal("No message was sent to user-3")
		}

		// Verify message content
		assert.Equal(t, api.MessageTypeNewConversation, msg1.Type)
		assert.Equal(t, api.MessageTypeNewConversation, msg2.Type)

		var payload1 api.WSNewConversationPayload
		err = json.Unmarshal(msg1.Data, &payload1)
		require.NoError(t, err)

		var payload2 api.WSNewConversationPayload
		err = json.Unmarshal(msg2.Data, &payload2)
		require.NoError(t, err)

		assert.Equal(t, req.ConversationID, payload1.ConversationID)
		assert.Equal(t, req.ConversationID, payload2.ConversationID)
		assert.Equal(t, senderID, payload1.SenderID)
		assert.Equal(t, senderID, payload2.SenderID)
		assert.Equal(t, []string{"user-1", "user-2", "user-3"}, payload1.ParticipantIDs)
		assert.Equal(t, []string{"user-1", "user-2", "user-3"}, payload2.ParticipantIDs)
		assert.Equal(t, req.OtherParticipants[0].KeyDistributionMessage, payload1.KeyDistributionMessage)
		assert.Equal(t, req.OtherParticipants[1].KeyDistributionMessage, payload2.KeyDistributionMessage)
	})

	t.Run("should store message for offline recipients", func(t *testing.T) {
		// Arrange
		db, dbClose := testDB(t)
		defer dbClose()

		convRepo := NewMockConversationRepository()
		manager := NewManager(db, convRepo)

		senderID := "user-1"
		req := api.NewConversationRequest{
			ConversationID: "conv-456",
			OtherParticipants: []api.Participant{
				{
					ID:                     "user-2",
					KeyDistributionMessage: []byte("key-distribution-message"),
				},
				{
					ID:                     "user-3",
					KeyDistributionMessage: []byte("key-distribution-message"),
				},
			},
		}

		// Act
		err := manager.BroadcastNewConversation(senderID, req)
		require.NoError(t, err)

		// Assert
		// Check that messages were stored for offline recipients
		messageStore1 := &MessageStore{
			db:       db,
			clientID: "user-2",
		}
		messages1, err := messageStore1.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages1, 1)
		assert.Equal(t, api.MessageTypeNewConversation, messages1[0].Type)

		var payload1 api.WSNewConversationPayload
		err = json.Unmarshal(messages1[0].Data, &payload1)
		require.NoError(t, err)
		assert.Equal(t, req.ConversationID, payload1.ConversationID)
		assert.Equal(t, senderID, payload1.SenderID)
		assert.Equal(t, []string{"user-1", "user-2", "user-3"}, payload1.ParticipantIDs)
		assert.Equal(t, req.OtherParticipants[0].KeyDistributionMessage, payload1.KeyDistributionMessage)

		messageStore2 := &MessageStore{
			db:       db,
			clientID: "user-3",
		}
		messages2, err := messageStore2.LoadAll()
		require.NoError(t, err)
		require.Len(t, messages2, 1)
		assert.Equal(t, api.MessageTypeNewConversation, messages2[0].Type)

		var payload2 api.WSNewConversationPayload
		err = json.Unmarshal(messages2[0].Data, &payload2)
		require.NoError(t, err)
		assert.Equal(t, req.ConversationID, payload2.ConversationID)
		assert.Equal(t, senderID, payload2.SenderID)
		assert.Equal(t, []string{"user-1", "user-2", "user-3"}, payload2.ParticipantIDs)
		assert.Equal(t, req.OtherParticipants[1].KeyDistributionMessage, payload2.KeyDistributionMessage)
	})
}
