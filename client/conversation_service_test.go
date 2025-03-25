package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/client/models"
	"signal-chat/internal/api"
	"strings"
	"testing"
	"time"
)

const (
	DummyValue = "dummy"
)

func TestConversationService_WebsocketHandlers(t *testing.T) {
	t.Run("Sync websocket message handler creates all pending conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		svc := NewConversationService(db, ac, encryption.NewManagerFake())

		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				RecipientIDs:   []string{"alice", "bob"},
			}},
		}
		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		expected := syncData.NewConversations[0]
		assert.Equal(t, expected.ConversationID, conversations[0].ID)
		assert.Equal(t, expected.RecipientIDs, conversations[0].RecipientIDs)
	})
	t.Run("Sync websocket message handler creates all pending messages and updates corresponding conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		en := encryption.NewManagerFake()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))
		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				RecipientIDs:   []string{"alice"},
			}},
			NewMessages: []api.WSNewMessagePayload{{
				ConversationID:   "123",
				MessageID:        "def",
				SenderID:         "alice",
				EncryptedMessage: encrypted.Serialized,
				Timestamp:        time.Now().UnixMilli(),
			}},
		}
		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, syncData.NewMessages[0].SenderID, conversations[0].LastMessageSenderID, "last message sender ID should have been updated to the sender ID of the last message")
		assert.True(t, strings.HasPrefix(text, conversations[0].LastMessagePreview), "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, syncData.NewMessages[0].Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(syncData.NewMessages[0].ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, syncData.NewMessages[0].MessageID, messages[0].ID)
		assert.Equal(t, syncData.NewMessages[0].SenderID, messages[0].SenderID)
		assert.Equal(t, syncData.NewMessages[0].Timestamp, messages[0].Timestamp)
		assert.Equal(t, text, messages[0].Text)
	})
	t.Run("Sync websocket message handler returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		expectedErr := errors.New("write error")
		db.WriteErr = expectedErr
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac, encryption.NewManagerFake())

		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: DummyValue,
				RecipientIDs:   []string{DummyValue},
			}},
		}
		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.Error(t, err, "websocket handler should have returned an error")
		assert.ErrorIs(t, err, expectedErr, "should pass database error to handler")
	})
	t.Run("NewConversation websocket message handler creates new conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		svc := NewConversationService(db, ac, encryption.NewManagerFake())

		payload := api.WSNewConversationPayload{
			ConversationID: "123",
			RecipientIDs:   []string{"alice", "bob"},
		}
		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeNewConversation,
			Data: mustMarshal(payload),
		}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, payload.ConversationID, conversations[0].ID)
		assert.Equal(t, payload.RecipientIDs, conversations[0].RecipientIDs)
	})
	t.Run("NewConversation websocket message handler invokes new conversation callback", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		svc := NewConversationService(db, ac, encryption.NewManagerFake())
		payload := api.WSNewConversationPayload{
			ConversationID: "123",
			RecipientIDs:   []string{"alice", "bob"},
		}
		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeNewConversation,
			Data: mustMarshal(payload),
		}}
		callback := false
		svc.ConversationAdded = func(conv models.Conversation) {
			callback = true
		}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		assert.True(t, callback, "new conversation callback should have been invoked")
	})
	t.Run("NewConversation websocket message handler returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		expectedErr := errors.New("write error")
		db.WriteErr = expectedErr
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac, encryption.NewManagerFake())

		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeNewConversation,
			Data: mustMarshal(api.WSNewConversationPayload{
				ConversationID: DummyValue,
				RecipientIDs:   []string{DummyValue},
			}),
		}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.Error(t, err, "websocket handler should have returned an error")
		assert.ErrorIs(t, err, expectedErr, "should pass database error to handler")
	})
	t.Run("NewMessage websocket message handler creates new messages and updates corresponding conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		en := encryption.NewManagerFake()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))
		newMessagePayload := api.WSNewMessagePayload{
			ConversationID:   "123",
			MessageID:        "def",
			SenderID:         "alice",
			EncryptedMessage: encrypted.Serialized,
			Timestamp:        time.Now().UnixMilli(),
		}
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewConversation,
				Data: mustMarshal(api.WSNewConversationPayload{
					ConversationID: "123",
					RecipientIDs:   []string{"bob"},
				}),
			},
			{
				Type: api.MessageTypeNewMessage,
				Data: mustMarshal(newMessagePayload),
			}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, newMessagePayload.SenderID, conversations[0].LastMessageSenderID, "last message sender ID should have been updated to the sender ID of the last message")
		assert.True(t, strings.HasPrefix(text, conversations[0].LastMessagePreview), "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, newMessagePayload.Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(newMessagePayload.ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, newMessagePayload.MessageID, messages[0].ID)
		assert.Equal(t, newMessagePayload.SenderID, messages[0].SenderID)
		assert.Equal(t, newMessagePayload.Timestamp, messages[0].Timestamp)
		assert.Equal(t, text, messages[0].Text)
	})
	t.Run("NewMessage websocket message handler creates invokes new message and updated conversation callbacks", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		en := encryption.NewManagerFake()
		svc := NewConversationService(db, ac, en)
		text := "Hello world!"
		encrypted, _ := en.GroupEncrypt("123", []byte(text))
		newMessagePayload := api.WSNewMessagePayload{
			ConversationID:   "123",
			MessageID:        "def",
			SenderID:         "alice",
			EncryptedMessage: encrypted.Serialized,
			Timestamp:        time.Now().UnixMilli(),
		}
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewConversation,
				Data: mustMarshal(api.WSNewConversationPayload{
					ConversationID: "123",
					RecipientIDs:   []string{"bob"},
				}),
			},
			{
				Type: api.MessageTypeNewMessage,
				Data: mustMarshal(newMessagePayload),
			}}

		msgCallback := false
		svc.MessageAdded = func(msg models.Message) {
			msgCallback = true
		}
		convCallback := false
		svc.ConversationUpdated = func(conv models.Conversation) {
			convCallback = true
		}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.NoError(t, err)
		assert.True(t, msgCallback, "new message callback should have been invoked")
		assert.True(t, convCallback, "updated conversation callback should have been invoked")
	})
	t.Run("NewMessage websocket message handler returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac, encryption.NewManagerFake())
		payload := api.WSNewMessagePayload{
			ConversationID:   "123",
			MessageID:        DummyValue,
			SenderID:         DummyValue,
			EncryptedMessage: []byte(DummyValue),
			Timestamp:        time.Now().UnixMilli(),
		}
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewMessage,
				Data: mustMarshal(payload),
			}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.Error(t, err, "websocket handler should have returned an error")
	})
	t.Run("NewMessage websocket message handler returns error when fails to read corresponding conversation from database", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.ReadErr = errors.New("read error")
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac, encryption.NewManagerFake())
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewConversation,
				Data: mustMarshal(api.WSNewConversationPayload{
					ConversationID: "123",
					RecipientIDs:   []string{DummyValue},
				}),
			},
			{
				Type: api.MessageTypeNewMessage,
				Data: mustMarshal(api.WSNewMessagePayload{
					ConversationID:   "123",
					MessageID:        DummyValue,
					SenderID:         DummyValue,
					EncryptedMessage: []byte(DummyValue),
					Timestamp:        time.Now().UnixMilli(),
				}),
			}}

		// Act
		err := ac.TriggerWebsocketMessages(wsMessages)

		// Assert
		assert.Error(t, err, "websocket handler should have returned an error")
		assert.ErrorIs(t, err, db.ReadErr, "should pass database error to handler")
	})
}

func TestConversationService_ListConversations(t *testing.T) {
	t.Run("returns all existing conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth(), encryption.NewManagerFake())
		conv1, err := svc.CreateConversation([]string{"bob"})
		require.NoError(t, err)
		conv2, err := svc.CreateConversation([]string{"tom"})
		require.NoError(t, err)
		conv3, err := svc.CreateConversation([]string{"alice"})
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
		service := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

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
		service := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{QueryErr: errors.New("query err")}
		service := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

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
		ac := apiclient.NewStub()
		resp := api.CreateConversationResponse{}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		svc := NewConversationService(db, ac, encryption.NewManagerFake())

		// Act
		conv, err := svc.CreateConversation([]string{"alice"})

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, conv.ID, "Conversation ID must be set")
		assert.Len(t, conv.RecipientIDs, 1, "conversation should have only one participant, conversation creator should not be included")
		assert.Contains(t, conv.RecipientIDs, "alice")
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Contains(t, conversations, conv, "conversation should be retrievable after creation")
	})
	t.Run("returns error if API client fails to send request", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		ac.PostErrors[api.EndpointConversations] = errors.New("test error")
		svc := NewConversationService(db, ac, encryption.NewManagerFake())

		// Act
		_, err := svc.CreateConversation([]string{DummyValue})

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error if server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       mustMarshal(api.CreateConversationResponse{Error: "test error"}),
		}
		svc := NewConversationService(db, ac, encryption.NewManagerFake())

		// Act
		_, err := svc.CreateConversation([]string{DummyValue})

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty recipientID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation([]string{}) })
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{WriteErr: errors.New("write err")}
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := svc.CreateConversation([]string{DummyValue})

		// Assert
		assert.Error(t, err)
	})
}

func TestConversationService_SendMessage(t *testing.T) {
	t.Run("creates a new message on successful response from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		resp := api.CreateMessageResponse{
			MessageID: "123",
			Timestamp: time.Now().UnixMilli(),
		}
		ac.PostResponses[api.EndpointMessages] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(api.CreateConversationResponse{}),
		}
		svc := NewConversationService(db, ac, encryption.NewManagerFake())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, resp.MessageID, msg.ID, "message ID returned from server should have been used")
		assert.Equal(t, "", msg.SenderID, "sender ID should be empty since it was us who sent the message")
		assert.Equal(t, resp.Timestamp, msg.Timestamp, "timestamp returned from server should have been used")
		assert.Equal(t, "Second message", msg.Text)
		messages, err := svc.ListMessages(conv.ID)
		require.NoError(t, err)
		assert.Contains(t, messages, msg, "message should be retrievable after creation")
	})
	t.Run("updates conversation and invokes updated conversation callback", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth(), encryption.NewManagerFake())
		conv, err := svc.CreateConversation([]string{DummyValue})
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
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := svc.SendMessage("123", DummyValue)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error if API client fails to send request", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(api.CreateConversationResponse{}),
		}
		ac.PostErrors[api.EndpointMessages] = errors.New("test error") // fail on create message request
		svc := NewConversationService(db, ac, encryption.NewManagerFake())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)

		// Act
		_, err = svc.SendMessage(conv.ID, DummyValue)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error if server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(api.CreateConversationResponse{}),
		}
		ac.PostResponses[api.EndpointMessages] = apiclient.StubResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       mustMarshal(api.CreateConversationResponse{Error: "test error"}),
		}
		svc := NewConversationService(db, ac, encryption.NewManagerFake())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)

		// Act
		_, err = svc.SendMessage(conv.ID, DummyValue)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		id := "123"
		key := string(database.ConversationPK(id))
		bytes, _ := (&models.Conversation{ID: id}).Serialize()
		db.QueryResult = map[string][]byte{key: bytes}
		db.WriteErr = errors.New("write error")
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := svc.SendMessage(id, "New message")

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty conversationID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		assert.Panics(t, func() { _, _ = svc.SendMessage("", "Test message") })
	})
	t.Run("panics when empty messageText", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		assert.Panics(t, func() { _, _ = svc.SendMessage("123", "") })
	})
}

func TestConversationService_ListMessages(t *testing.T) {
	t.Run("returns all messages from the given conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth(), encryption.NewManagerFake())
		conv, err := svc.CreateConversation([]string{DummyValue})
		require.NoError(t, err)
		msg, err := svc.SendMessage(conv.ID, DummyValue)
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
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = svc.ListMessages("")
		})
	})
	t.Run("returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		messages, err := svc.ListMessages("123")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, messages)
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
		svc := NewConversationService(db, apiclient.NewFake(), encryption.NewManagerFake())
		require.NoError(t, err)

		// Act
		got, err := svc.ListMessages(conv.ID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return b
}
