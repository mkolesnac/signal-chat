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
	"signal-chat/client/models"
	"signal-chat/internal/api"
	"testing"
)

const (
	DummyValue = "dummy"
)

func TestNewConversationService(t *testing.T) {
	t.Run("processes all pending conversations from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				ParticipantIDs: []string{"alice", "bob"},
				SenderID:       "alice",
				MessageID:      "abc",
				MessageText:    "First message!",
				MessagePreview: "First...",
				Timestamp:      "2024-02-16T10:00:00Z",
			}},
		}
		ac.WSMessages = []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		svc := NewConversationService(db, ac)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		expected := syncData.NewConversations[0]
		assert.Equal(t, expected.ConversationID, conversations[0].ID)
		assert.Equal(t, expected.ParticipantIDs, conversations[0].ParticipantIDs)
		assert.Equal(t, expected.SenderID, conversations[0].LastMessageSenderID)
		assert.Equal(t, expected.MessagePreview, conversations[0].LastMessagePreview)
		assert.Equal(t, expected.Timestamp, conversations[0].LastMessageTimestamp)
		messages, err := svc.ListMessages(expected.ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "One message should have been created")
		assert.Equal(t, expected.MessageID, messages[0].ID)
		assert.Equal(t, expected.SenderID, messages[0].SenderID)
		assert.Equal(t, expected.MessageText, messages[0].Text)
		assert.Equal(t, expected.Timestamp, messages[0].Timestamp)
	})
	t.Run("processes all pending messages from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				ParticipantIDs: []string{"alice", "bob"},
				SenderID:       "alice",
				MessageID:      "abc",
				MessageText:    "First message!",
				MessagePreview: "First...",
				Timestamp:      "2024-02-16T10:00:00Z",
			}},
			NewMessages: []api.WSNewMessagePayload{{
				ConversationID: "123",
				MessageID:      "def",
				SenderID:       "bob",
				Text:           "Second message!",
				Preview:        "Second...",
				Timestamp:      "2024-03-16T10:00:00Z",
			}},
		}
		ac.WSMessages = []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		svc := NewConversationService(db, ac)

		// Assert
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Len(t, conversations, 1, "One conversation should have been created")
		assert.Equal(t, syncData.NewMessages[0].SenderID, conversations[0].LastMessageSenderID, "last message sender ID should have been updated to the sender ID of the last message")
		assert.Equal(t, syncData.NewMessages[0].Preview, conversations[0].LastMessagePreview, "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, syncData.NewMessages[0].Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(syncData.NewMessages[0].ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 2, "Two message should have been created")
		assert.Equal(t, syncData.NewMessages[0].MessageID, messages[1].ID)
		assert.Equal(t, syncData.NewMessages[0].SenderID, messages[1].SenderID)
		assert.Equal(t, syncData.NewMessages[0].Text, messages[1].Text)
		assert.Equal(t, syncData.NewMessages[0].Timestamp, messages[1].Timestamp)
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		expectedErr := errors.New("write error")
		db.WriteErr = expectedErr
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()

		var capturedErr error
		ac.SetErrorHandler(func(err error) {
			capturedErr = err
		})

		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				ParticipantIDs: []string{"alice", "bob"},
				SenderID:       "alice",
				MessageID:      "abc",
				MessageText:    "First message!",
				MessagePreview: "First...",
				Timestamp:      "2024-02-16T10:00:00Z",
			}},
		}
		ac.WSMessages = []api.WSMessage{{
			Type: api.MessageTypeSync,
			Data: mustMarshal(syncData),
		}}

		// Act
		_ = NewConversationService(db, ac)

		// Assert
		assert.Error(t, capturedErr, "error handler on API client should have been called")
		assert.ErrorIs(t, capturedErr, expectedErr, "should pass database error to handler")
	})
}

func TestConversationService_ListConversations(t *testing.T) {
	t.Run("returns all existing conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth())
		conv1, err := svc.CreateConversation("First message", "bob")
		require.NoError(t, err)
		conv2, err := svc.CreateConversation("First message", "tom")
		require.NoError(t, err)

		// Act
		conversations, err := svc.ListConversations()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, conversations, 2)
		assert.Contains(t, conversations, conv1)
		assert.Contains(t, conversations, conv2)
	})
	t.Run("returns empty list when no conversations exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		service := NewConversationService(db, apiclient.NewFake())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when unparsable conversation in database", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		_ = db.Write(database.ConversationPK("abc"), []byte("invalid"))
		service := NewConversationService(db, apiclient.NewFake())

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{QueryErr: errors.New("query err")}
		service := NewConversationService(db, apiclient.NewFake())

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
		resp := api.CreateConversationResponse{
			ConversationID: "123",
			MessageID:      DummyValue,
			SenderID:       "current-user",
			ParticipantIDs: []string{"current-user", "bob"},
			Timestamp:      "2024-02-16T10:00:00Z",
		}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		svc := NewConversationService(db, ac)
		messageText := "Hello there!"

		// Act
		conv, err := svc.CreateConversation(messageText, "bob")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, resp.ConversationID, conv.ID, "server's conversation ID must be preserved")
		assert.Equal(t, resp.ParticipantIDs, conv.ParticipantIDs, "server's participant IDs must be preserved")
		assert.Equal(t, resp.Timestamp, conv.LastMessageTimestamp, "last message timestamp should use value returned from server")
		assert.Equal(t, resp.SenderID, conv.LastMessageSenderID, "server's sender ID must be preserved")
		assert.Contains(t, messageText, conv.LastMessagePreview, "generated message preview should be part of the original message text")
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		assert.Contains(t, conversations, conv, "conversation should be retrievable after creation")
	})
	t.Run("creates initial conversation message", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		resp := api.CreateConversationResponse{
			ConversationID: DummyValue,
			MessageID:      "msg#123",
			SenderID:       "current-user",
			ParticipantIDs: []string{"current-user", "bob"},
			Timestamp:      "2024-02-16T10:00:00Z",
		}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		svc := NewConversationService(db, ac)
		messageText := "Hello there!"

		// Act
		conv, err := svc.CreateConversation(messageText, DummyValue)

		// Assert
		assert.NoError(t, err)
		messages, err := svc.ListMessages(conv.ID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "only one message should be created")
		assert.Equal(t, messageText, messages[0].Text, "a message with matching text should have been created")
		assert.Equal(t, resp.Timestamp, messages[0].Timestamp, "timestamp returned from the server should have been used")
		assert.Equal(t, resp.SenderID, messages[0].SenderID, "sender ID returned from the server should have been used")
	})
	t.Run("returns error if API client fails to send request", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		ac.PostErrors[api.EndpointConversations] = errors.New("test error")
		svc := NewConversationService(db, ac)

		// Act
		_, err := svc.CreateConversation(DummyValue, DummyValue)

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
		svc := NewConversationService(db, ac)

		// Act
		_, err := svc.CreateConversation(DummyValue, DummyValue)

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty messageText", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation("", "bob") })
	})
	t.Run("panics when empty recipientID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation("test", "") })
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{WriteErr: errors.New("write err")}
		svc := NewConversationService(db, apiclient.NewFake())

		// Act
		_, err := svc.CreateConversation("Initial message", "bob")

		// Assert
		assert.Error(t, err)
	})
}

func TestConversationService_SendMessage(t *testing.T) {
	t.Run("creates a new message on successful response from server", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth())
		conv, err := svc.CreateConversation(DummyValue, DummyValue)
		require.NoError(t, err)

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.ID, "message ID returned from server should have been used")
		assert.NotEmpty(t, msg.SenderID, "sender ID returned from server should have been used")
		assert.NotEmpty(t, msg.Timestamp, "timestamp returned from server should have been used")
		assert.Equal(t, "Second message", msg.Text)
		messages, err := svc.ListMessages(conv.ID)
		require.NoError(t, err)
		assert.Contains(t, messages, msg, "message should be retrievable after creation")
	})
	t.Run("updates conversation's last message preview", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth())
		conv, err := svc.CreateConversation(DummyValue, DummyValue)
		require.NoError(t, err)

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		conversations, err := svc.ListConversations()
		require.NoError(t, err)
		require.Len(t, conversations, 1)
		assert.Contains(t, msg.Text, conversations[0].LastMessagePreview, "last message preview in conversation should have been updated")
		assert.Equal(t, msg.Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp in conversation should have been updated")
		assert.Equal(t, msg.SenderID, conversations[0].LastMessageSenderID, "last message sender ID in conversation should have been updated")
	})
	t.Run("returns error when given conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFake())

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
			Body:       mustMarshal(api.CreateConversationResponse{ConversationID: "123"}),
		}
		ac.PostErrors[api.EndpointMessages] = errors.New("test error") // fail on create message request
		svc := NewConversationService(db, ac)
		conv, err := svc.CreateConversation(DummyValue, DummyValue)
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
			Body:       mustMarshal(api.CreateConversationResponse{ConversationID: "123"}),
		}
		ac.PostResponses[api.EndpointMessages] = apiclient.StubResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       mustMarshal(api.CreateConversationResponse{Error: "test error"}),
		}
		svc := NewConversationService(db, ac)
		conv, err := svc.CreateConversation(DummyValue, DummyValue)
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
		svc := NewConversationService(db, apiclient.NewFake())

		// Act
		_, err := svc.SendMessage(id, "New message")

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty conversationID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation("", "New message") })
	})
	t.Run("panics when empty messageText", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.SendMessage("123", "") })
	})
}

func TestConversationService_ListMessages(t *testing.T) {
	t.Run("returns all messages from the given conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth())
		conv, err := svc.CreateConversation(DummyValue, DummyValue)
		if err != nil {
			t.Fatal(err)
		}
		msg, err := svc.SendMessage(conv.ID, DummyValue)
		if err != nil {
			t.Fatal(err)
		}

		// Act
		messages, err := svc.ListMessages(conv.ID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Contains(t, messages, msg)
	})
	t.Run("panics when empty conversation ID", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFake())

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = svc.ListMessages("")
		})
	})
	t.Run("returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		svc := NewConversationService(db, apiclient.NewFake())

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
		svc := NewConversationService(db, apiclient.NewFake())
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
