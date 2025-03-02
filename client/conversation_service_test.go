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
		svc := NewConversationService(db, ac)

		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				Name:           "Test conversation",
				ParticipantIDs: []string{"alice", "bob"},
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
		assert.Equal(t, expected.Name, conversations[0].Name)
		assert.Equal(t, expected.ParticipantIDs, conversations[0].ParticipantIDs)
	})
	t.Run("Sync websocket message handler creates all pending messages and updates corresponding conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		svc := NewConversationService(db, ac)
		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: "123",
				Name:           "Test conversation",
				ParticipantIDs: []string{"mer", "bob"},
			}},
			NewMessages: []api.WSNewMessagePayload{{
				ConversationID: "123",
				MessageID:      "def",
				SenderID:       "bob",
				Text:           "Hello world!",
				Preview:        "Hello...",
				Timestamp:      time.Now().UnixMilli(),
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
		assert.Equal(t, syncData.NewMessages[0].Preview, conversations[0].LastMessagePreview, "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, syncData.NewMessages[0].Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(syncData.NewMessages[0].ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, syncData.NewMessages[0].MessageID, messages[0].ID)
		assert.Equal(t, syncData.NewMessages[0].SenderID, messages[0].SenderID)
		assert.Equal(t, syncData.NewMessages[0].Text, messages[0].Text)
		assert.Equal(t, syncData.NewMessages[0].Timestamp, messages[0].Timestamp)
	})
	t.Run("Sync websocket message handler returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		expectedErr := errors.New("write error")
		db.WriteErr = expectedErr
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac)

		syncData := api.WSSyncData{
			NewConversations: []api.WSNewConversationPayload{{
				ConversationID: DummyValue,
				Name:           DummyValue,
				ParticipantIDs: []string{DummyValue},
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
		svc := NewConversationService(db, ac)

		payload := api.WSNewConversationPayload{
			ConversationID: "123",
			Name:           "Test conversation",
			ParticipantIDs: []string{"alice", "bob"},
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
		assert.Equal(t, payload.Name, conversations[0].Name)
		assert.Equal(t, payload.ParticipantIDs, conversations[0].ParticipantIDs)
	})
	t.Run("NewConversation websocket message handler returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		expectedErr := errors.New("write error")
		db.WriteErr = expectedErr
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac)

		wsMessages := []api.WSMessage{{
			Type: api.MessageTypeNewConversation,
			Data: mustMarshal(api.WSNewConversationPayload{
				ConversationID: DummyValue,
				Name:           DummyValue,
				ParticipantIDs: []string{DummyValue},
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
		svc := NewConversationService(db, ac)

		newMessagePayload := api.WSNewMessagePayload{
			ConversationID: "123",
			MessageID:      "def",
			SenderID:       "bob",
			Text:           "Hello world!",
			Preview:        "Hello...",
			Timestamp:      time.Now().UnixMilli(),
		}
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewConversation,
				Data: mustMarshal(api.WSNewConversationPayload{
					ConversationID: "123",
					Name:           DummyValue,
					ParticipantIDs: []string{"alice", "bob"},
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
		assert.Equal(t, newMessagePayload.Preview, conversations[0].LastMessagePreview, "last message preview should have been updated to the preview of the last message")
		assert.Equal(t, newMessagePayload.Timestamp, conversations[0].LastMessageTimestamp, "last message timestamp should have been updated to the timestamp of the last message")
		messages, err := svc.ListMessages(newMessagePayload.ConversationID)
		require.NoError(t, err)
		assert.Len(t, messages, 1, "A message should have been created")
		assert.Equal(t, newMessagePayload.MessageID, messages[0].ID)
		assert.Equal(t, newMessagePayload.SenderID, messages[0].SenderID)
		assert.Equal(t, newMessagePayload.Text, messages[0].Text)
		assert.Equal(t, newMessagePayload.Timestamp, messages[0].Timestamp)
	})
	t.Run("NewMessage websocket message handler returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyValue)
		ac := apiclient.NewStub()
		_ = NewConversationService(db, ac)

		payload := api.WSNewMessagePayload{
			ConversationID: "123",
			MessageID:      DummyValue,
			SenderID:       DummyValue,
			Text:           DummyValue,
			Preview:        DummyValue,
			Timestamp:      time.Now().UnixMilli(),
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
		_ = NewConversationService(db, ac)
		wsMessages := []api.WSMessage{
			{
				Type: api.MessageTypeNewConversation,
				Data: mustMarshal(api.WSNewConversationPayload{
					ConversationID: "123",
					Name:           DummyValue,
					ParticipantIDs: []string{DummyValue},
				}),
			},
			{
				Type: api.MessageTypeNewMessage,
				Data: mustMarshal(api.WSNewMessagePayload{
					ConversationID: "123",
					MessageID:      DummyValue,
					SenderID:       DummyValue,
					Text:           DummyValue,
					Preview:        DummyValue,
					Timestamp:      time.Now().UnixMilli(),
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
		svc := NewConversationService(db, apiclient.NewFakeWithoutAuth())
		conv1, err := svc.CreateConversation("Conversation1", []string{"bob"})
		require.NoError(t, err)
		conv2, err := svc.CreateConversation("Conversation2", []string{"tom"})
		require.NoError(t, err)
		conv3, err := svc.CreateConversation("Conversation3", []string{"alice"})
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
			ParticipantIDs: []string{"me", "alice"},
		}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		svc := NewConversationService(db, ac)

		// Act
		conv, err := svc.CreateConversation("Test conversation", []string{"alice"})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, resp.ConversationID, conv.ID, "server's conversation ID must be preserved")
		assert.Equal(t, "Test conversation", conv.Name, "conversation name must be set")
		assert.Equal(t, resp.ParticipantIDs, conv.ParticipantIDs, "server's participant IDs must be preserved")
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
		svc := NewConversationService(db, ac)

		// Act
		_, err := svc.CreateConversation(DummyValue, []string{DummyValue})

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
		_, err := svc.CreateConversation(DummyValue, []string{DummyValue})

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty messageText", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation("", []string{DummyValue}) })
	})
	t.Run("panics when empty recipientID", func(t *testing.T) {
		db := database.NewFake()
		err := db.Open(DummyValue)
		require.NoError(t, err)
		svc := NewConversationService(db, apiclient.NewFake())

		assert.Panics(t, func() { _, _ = svc.CreateConversation(DummyValue, []string{}) })
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{WriteErr: errors.New("write err")}
		svc := NewConversationService(db, apiclient.NewFake())

		// Act
		_, err := svc.CreateConversation(DummyValue, []string{DummyValue})

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
			SenderID:  "465",
			Timestamp: time.Now().UnixMilli(),
		}
		ac.PostResponses[api.EndpointMessages] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       mustMarshal(resp),
		}
		ac.PostResponses[api.EndpointConversations] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body: mustMarshal(api.CreateConversationResponse{
				ConversationID: DummyValue,
				ParticipantIDs: []string{DummyValue},
			}),
		}
		svc := NewConversationService(db, ac)
		conv, err := svc.CreateConversation(DummyValue, []string{DummyValue})
		require.NoError(t, err)

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, resp.MessageID, msg.ID, "message ID returned from server should have been used")
		assert.Equal(t, resp.SenderID, msg.SenderID, "sender ID returned from server should have been used")
		assert.Equal(t, resp.Timestamp, msg.Timestamp, "timestamp returned from server should have been used")
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
		conv, err := svc.CreateConversation(DummyValue, []string{DummyValue})
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
		conv, err := svc.CreateConversation(DummyValue, []string{DummyValue})
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
		conv, err := svc.CreateConversation(DummyValue, []string{DummyValue})
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

		assert.Panics(t, func() { _, _ = svc.SendMessage("", "Test message") })
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
		conv, err := svc.CreateConversation(DummyValue, []string{DummyValue})
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
