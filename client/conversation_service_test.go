package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"signal-chat/client/api"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/client/models"
	"signal-chat/internal/apitypes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DummyValue = "dummy"
)

func TestConversationService_WebsocketHandlers(t *testing.T) {
	t.Run("Sync websocket message handler creates all pending conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		svc := NewConversationService(db, ac, encryption.NewFakeManager())

		convPayload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"alice", "bob"},
			SenderID:               "sender1",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}
		syncPayload := apitypes.WSSyncPayload{
			Messages: []apitypes.WSMessage{{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewConversation,
				Data: mustMarshal(convPayload),
			}},
		}
		wsMessages := []apitypes.WSMessage{{
			Type: apitypes.MessageTypeSync,
			Data: mustMarshal(syncPayload),
		}}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, convPayload.ConversationID, conversations[0].ID)
		assert.Equal(t, convPayload.ParticipantIDs, conversations[0].ParticipantIDs)
	})

	t.Run("Sync websocket message handler creates all pending messages and updates corresponding conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		en := encryption.NewFakeManager()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))
		convPayload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"alice", "bob"},
			SenderID:               "alice",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}
		newMsgPayload := apitypes.WSNewMessagePayload{
			ConversationID: "123",
			MessageID:      "def",
			SenderID:       "alice",
			Content:        encrypted.Serialized,
			CreatedAt:      time.Now().UnixMilli(),
		}
		syncPayload := apitypes.WSSyncPayload{
			Messages: []apitypes.WSMessage{
				{
					ID:   "msg1",
					Type: apitypes.MessageTypeNewConversation,
					Data: mustMarshal(convPayload),
				},
				{
					ID:   "msg2",
					Type: apitypes.MessageTypeNewMessage,
					Data: mustMarshal(newMsgPayload),
				},
			},
		}
		wsMessages := []apitypes.WSMessage{{
			Type: apitypes.MessageTypeSync,
			Data: mustMarshal(syncPayload),
		}}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, newMsgPayload.SenderID, conversations[0].LastMessageSenderID, "last message sender MessageID should have been updated to the sender MessageID of the last message")
		assert.True(t, strings.HasPrefix(text, conversations[0].LastMessagePreview), "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, newMsgPayload.CreatedAt, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(newMsgPayload.ConversationID)
		require.NoError(t, err)
		require.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, newMsgPayload.MessageID, messages[0].ID)
		assert.Equal(t, newMsgPayload.SenderID, messages[0].SenderID)
		assert.Equal(t, newMsgPayload.CreatedAt, messages[0].Timestamp)
		assert.Equal(t, text, messages[0].Text)
	})

	t.Run("NewConversation websocket message handler creates new conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		en := encryption.NewFakeManager()
		svc := NewConversationService(db, ac, en)

		payload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"alice", "bob"},
			SenderID:               "sender1",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}
		wsMessages := []apitypes.WSMessage{{
			ID:   "msg1",
			Type: apitypes.MessageTypeNewConversation,
			Data: mustMarshal(payload),
		}}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, payload.ConversationID, conversations[0].ID)
		assert.Equal(t, payload.ParticipantIDs, conversations[0].ParticipantIDs)
	})

	t.Run("NewConversation websocket message handler invokes new conversation callback", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		en := encryption.NewFakeManager()
		svc := NewConversationService(db, ac, en)
		payload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"alice", "bob"},
			SenderID:               "sender1",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}
		wsMessages := []apitypes.WSMessage{{
			ID:   "msg1",
			Type: apitypes.MessageTypeNewConversation,
			Data: mustMarshal(payload),
		}}
		callback := false
		svc.ConversationAdded = func(conv models.Conversation) {
			callback = true
		}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.True(t, callback, "new conversation callback should have been invoked")
	})

	t.Run("NewMessage websocket message handler creates new messages and updates corresponding conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		en := encryption.NewFakeManager()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))

		// First create a conversation
		convPayload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"alice"},
			SenderID:               "alice",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}

		// Then create a message for it
		msgPayload := apitypes.WSNewMessagePayload{
			ConversationID: "123",
			MessageID:      "def",
			SenderID:       "alice",
			Content:        encrypted.Serialized,
			CreatedAt:      time.Now().UnixMilli(),
		}

		wsMessages := []apitypes.WSMessage{
			{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewConversation,
				Data: mustMarshal(convPayload),
			},
			{
				ID:   "msg2",
				Type: apitypes.MessageTypeNewMessage,
				Data: mustMarshal(msgPayload),
			},
		}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, msgPayload.SenderID, conversations[0].LastMessageSenderID, "last message sender ID should have been updated to the sender ID of the last message")
		assert.True(t, strings.HasPrefix(conversations[0].LastMessagePreview, "Hello world"), "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, msgPayload.CreatedAt, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")

		messages, err := svc.ListMessages(msgPayload.ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, msgPayload.MessageID, messages[0].ID)
		assert.Equal(t, msgPayload.SenderID, messages[0].SenderID)
		assert.Equal(t, msgPayload.CreatedAt, messages[0].Timestamp)
		assert.Contains(t, messages[0].Text, "Hello world")
	})

	t.Run("NewMessage websocket message handler invokes new message and updated conversation callbacks", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		en := encryption.NewFakeManager()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))

		// First create a conversation
		convPayload := apitypes.WSNewConversationPayload{
			ConversationID:         "123",
			ParticipantIDs:         []string{"bob"},
			SenderID:               "alice",
			KeyDistributionMessage: []byte("key-distribution-message"),
		}

		// Then create a message for it
		msgPayload := apitypes.WSNewMessagePayload{
			ConversationID: "123",
			MessageID:      "def",
			SenderID:       "alice",
			Content:        encrypted.Serialized,
			CreatedAt:      time.Now().UnixMilli(),
		}

		wsMessages := []apitypes.WSMessage{
			{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewConversation,
				Data: mustMarshal(convPayload),
			},
			{
				ID:   "msg2",
				Type: apitypes.MessageTypeNewMessage,
				Data: mustMarshal(msgPayload),
			},
		}

		msgCallback := false
		svc.MessageAdded = func(msg models.Message) {
			msgCallback = true
		}

		convCallback := false
		svc.ConversationUpdated = func(conv models.Conversation) {
			convCallback = true
		}

		// Act
		ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.True(t, msgCallback, "new message callback should have been invoked")
		assert.True(t, convCallback, "updated conversation callback should have been invoked")
	})
}

func TestConversationService_ListConversations(t *testing.T) {
	t.Run("returns all existing conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewFakeClient()

		// Create multiple users that will be used as conversation participants
		user1, _ := ac.SignUp("user1", "password", apitypes.KeyBundle{})
		user2, _ := ac.SignUp("user2", "password", apitypes.KeyBundle{})
		_, _ = ac.SignUp("me", "password", apitypes.KeyBundle{})

		svc := NewConversationService(db, ac, encryption.NewFakeManager())
		conv1, err := svc.CreateConversation([]string{user1.UserID})
		require.NoError(t, err)
		conv2, err := svc.CreateConversation([]string{user2.UserID})
		require.NoError(t, err)
		conv3, err := svc.CreateConversation([]string{user1.UserID, user2.UserID})
		require.NoError(t, err)

		// Act
		conversations, err := svc.ListConversations()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, conversations, 3)
		assert.Contains(t, conversations, conv1)
		assert.Contains(t, conversations, conv2)
		assert.Contains(t, conversations, conv3)
	})
	t.Run("returns empty list when no conversations exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		service := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when unparsable conversation in database", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.QueryResult = map[string][]byte{"conversation#123": []byte("invalid")}
		service := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{QueryErr: errors.New("query err")}
		service := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestConversationService_CreateConversation(t *testing.T) {
	t.Run("creates conversation on successful response from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		svc := NewConversationService(db, ac, encryption.NewFakeManager())

		// Act
		conv, err := svc.CreateConversation([]string{"alice"})

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, conv.ID, "Conversation ID must be set")
		assert.Len(t, conv.ParticipantIDs, 1, "conversation should have only one participant, conversation creator should not be included")
		assert.Contains(t, conv.ParticipantIDs, "alice")
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Contains(t, conversations, conv, "conversation should be retrievable after creation")
	})
	t.Run("returns error if API client fails to send request", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		ac.CreateConversationError = errors.New("test error")
		svc := NewConversationService(db, ac, encryption.NewFakeManager())

		// Act
		_, err := svc.CreateConversation([]string{DummyValue})

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create conversation")
	})
	t.Run("panics when empty recipientID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		assert.Panics(t, func() { _, _ = svc.CreateConversation([]string{}) })
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{WriteErr: errors.New("write err")}

		ac := api.NewFakeClient()
		user1, _ := ac.SignUp("user1", "password", apitypes.KeyBundle{})
		_, _ = ac.SignUp("me", "password", apitypes.KeyBundle{})

		svc := NewConversationService(db, ac, encryption.NewFakeManager())

		// Act
		_, err := svc.CreateConversation([]string{user1.UserID})

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store conversation")
	})
}

func TestConversationService_SendMessage(t *testing.T) {
	t.Run("creates a new message on successful response from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		resp := apitypes.SendMessageResponse{
			MessageID: "123",
			CreatedAt: time.Now().UnixMilli(),
		}
		ac.SendMessageResponse = resp
		svc := NewConversationService(db, ac, encryption.NewFakeManager())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, resp.MessageID, msg.ID, "message ID returned from server should have been used")
		assert.Equal(t, "", msg.SenderID, "sender ID should be empty since it was us who sent the message")
		assert.Equal(t, resp.CreatedAt, msg.Timestamp, "timestamp returned from server should have been used")
		assert.Equal(t, "Second message", msg.Text)
		messages, err := svc.ListMessages(conv.ID)
		require.NoError(t, err)
		assert.Contains(t, messages, msg, "message should be retrievable after creation")
	})

	t.Run("updates conversation and invokes updated conversation callback", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)

		ac := api.NewFakeClient()
		user1, _ := ac.SignUp("user1", "password", apitypes.KeyBundle{})
		_, _ = ac.SignUp("me", "password", apitypes.KeyBundle{})

		svc := NewConversationService(db, ac, encryption.NewFakeManager())
		conv, err := svc.CreateConversation([]string{user1.UserID})
		require.NoError(t, err)

		var updatedConv models.Conversation
		called := false
		svc.ConversationUpdated = func(conv models.Conversation) {
			called = true
			updatedConv = conv
		}

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		assert.True(t, called, "updated conversation callback should have been invoked")
		assert.Contains(t, msg.Text, updatedConv.LastMessagePreview, "last message preview in conversation should have been updated")
		assert.Equal(t, msg.Timestamp, updatedConv.LastMessageTimestamp, "last message timestamp in conversation should have been updated")
		assert.Equal(t, msg.SenderID, updatedConv.LastMessageSenderID, "last message sender ID in conversation should have been updated")
	})

	t.Run("returns error when given conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		_, err := svc.SendMessage("123", DummyValue)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conversation not found")
	})
	t.Run("returns error if API client fails to send request", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := api.NewStubClient()
		ac.SendMessageError = errors.New("test error") // fail on create message request
		svc := NewConversationService(db, ac, encryption.NewFakeManager())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)

		// Act
		_, err = svc.SendMessage(conv.ID, DummyValue)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send message")
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		id := "123"
		key := string(database.ConversationPK(id))
		bytes, _ := (&models.Conversation{ID: id}).Serialize()
		db.QueryResult = map[string][]byte{key: bytes}
		db.WriteErr = errors.New("write error")
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		_, err := svc.SendMessage(id, "New message")

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty conversationID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		assert.Panics(t, func() { _, _ = svc.SendMessage("", "Test message") })
	})
	t.Run("panics when empty messageText", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		assert.Panics(t, func() { _, _ = svc.SendMessage("123", "") })
	})
}

func TestConversationService_ListMessages(t *testing.T) {
	t.Run("returns all messages from the given conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)

		ac := api.NewFakeClient()
		user1, _ := ac.SignUp("user1", "password", apitypes.KeyBundle{})
		_, _ = ac.SignUp("me", "password", apitypes.KeyBundle{})

		svc := NewConversationService(db, ac, encryption.NewFakeManager())
		conv, err := svc.CreateConversation([]string{user1.UserID})
		require.NoError(t, err)

		msg, err := svc.SendMessage(conv.ID, "Test message")
		require.NoError(t, err)

		// Act
		messages, err := svc.ListMessages(conv.ID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Contains(t, messages, msg)
	})

	t.Run("panics when empty conversation ID", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = svc.ListMessages("")
		})
	})
	t.Run("returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		messages, err := svc.ListMessages("123")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, messages)
		assert.Contains(t, err.Error(), "conversation not found")
	})

	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		conv := models.Conversation{ID: "123"}
		bytes, err := conv.Serialize()
		require.NoError(t, err)
		db := &database.Stub{
			ReadResult: bytes,
			QueryErr:   errors.New("query err"),
		}
		svc := NewConversationService(db, api.NewFakeClient(), encryption.NewFakeManager())
		require.NoError(t, err)

		// Act
		got, err := svc.ListMessages(conv.ID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
		assert.Contains(t, err.Error(), "failed to query messages")
	})
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return b
}
