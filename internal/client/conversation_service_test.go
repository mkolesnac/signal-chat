package client

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"signal-chat/internal/client/database"
	"testing"
)

const DummyUser = "user#dummy"
const DummyText = "Hello there!"

func TestConversationService_ListConversations(t *testing.T) {
	t.Run("returns all existing conversations", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		conv1, err := svc.CreateConversation("First message", "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}
		conv2, err := svc.CreateConversation("First message", "user#1", "user#3")
		if err != nil {
			t.Fatal(err)
		}

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
		_ = db.Open(DummyUser)
		service := NewConversationService(db)

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when invalid conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		_ = db.WriteValue(database.ConversationPK("abc"), []byte("invalid"))
		service := NewConversationService(db)

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{QueryErr: errors.New("query err")}
		service := NewConversationService(db)

		// Act
		got, err := service.ListConversations()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}

func TestConversationService_CreateConversation(t *testing.T) {
	t.Run("creates conversation with given message text and sender ID", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		senderID := "user#1"
		recipientID := "user#2"
		messageText := "Hello there!"

		// Act
		conv, err := svc.CreateConversation(messageText, senderID, recipientID)

		// Assert
		assert.NoError(t, err)
		assert.Contains(t, messageText, conv.LastMessagePreview, "generated message preview should be part of the original message text")
		assert.Equal(t, senderID, conv.LastMessageSenderID)
		assert.Contains(t, conv.ParticipantIDs, senderID, "sender should be a conversation participant")
		assert.Contains(t, conv.ParticipantIDs, recipientID, "recipient should be a conversation participant")
		conversations, err := svc.ListConversations()
		assert.NoError(t, err)
		assert.Len(t, conversations, 1, "only one conversation should be created")
		assert.Contains(t, conversations, conv, "conversation should be retrievable after it was created")
	})
	t.Run("creates initial conversation message", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		senderID := "user#1"
		recipientID := "user#2"
		messageText := "Hello there!"

		// Act
		conv, _ := svc.CreateConversation(messageText, senderID, recipientID)

		// Assert
		messages, err := svc.ListMessages(conv.ID)
		assert.NoError(t, err)
		assert.Len(t, messages, 1, "only one message should be created")
		assert.Equal(t, messages[0].Text, messageText, "a message with matching text should have been created")
		assert.Equal(t, messages[0].SenderID, senderID)
	})
	t.Run("creates conversations with unique IDs", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)

		// Act
		conv1, err1 := svc.CreateConversation("First message", "user#1", "user#2")
		conv2, err2 := svc.CreateConversation("First message", "user#1", "user#3")

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEmpty(t, conv1.ID)
		assert.NotEmpty(t, conv2.ID)
		assert.NotEqual(t, conv1.ID, conv2.ID)
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{WriteErr: errors.New("write err")}
		svc := NewConversationService(db)

		// Act
		_, err := svc.CreateConversation("Initial message", "user#1", "user#2")

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics with empty required args", func(t *testing.T) {
		// Arrange
		tests := []struct {
			name        string
			messageText string
			senderID    string
			recipientID string
		}{
			{
				name:        "empty message text",
				messageText: "",
				senderID:    "user#1",
				recipientID: "user#2",
			},
			{
				name:        "empty sender ID",
				messageText: "Hello",
				senderID:    "",
				recipientID: "user#2",
			},
			{
				name:        "empty recipient ID",
				messageText: "Hello",
				senderID:    "user#1",
				recipientID: "",
			},
		}

		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Panics(t, func() {
					_, _ = svc.CreateConversation(tt.messageText, tt.senderID, tt.recipientID)
				})
			})
		}
	})
}

func TestConversationService_SendMessage(t *testing.T) {
	t.Run("adds new message to conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		conv, err := svc.CreateConversation(DummyText, "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		msg, err := svc.SendMessage(conv.ID, "Second message", "user#2")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Second message", msg.Text)
		assert.Equal(t, "user#2", msg.SenderID)
		assert.Equal(t, conv.ID, msg.ConversationID)
		messages, err := svc.ListMessages(conv.ID)
		assert.NoError(t, err)
		assert.Contains(t, messages, msg)
	})
	t.Run("creates messages with unique IDs", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		conv, err := svc.CreateConversation(DummyText, "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		msg1, err1 := svc.SendMessage(conv.ID, "First message", "user#2")
		msg2, err2 := svc.SendMessage(conv.ID, "Second message", "user#2")

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEmpty(t, msg1.ID)
		assert.NotEmpty(t, msg2.ID)
		assert.NotEqual(t, msg1.ID, msg2.ID)
	})
	t.Run("returns error when given conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)

		// Act
		_, err := svc.SendMessage("conv#123", "Second message", "user#2")

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when database write fails", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		id := "conversation#123"
		convBytes, _ := (&Conversation{ID: id}).Serialize()
		db.Items[database.ConversationPK(id)] = convBytes
		db.WriteErr = errors.New("write error")
		svc := NewConversationService(db)

		// Act
		_, err := svc.SendMessage(id, "Initial message", "user#1")

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics with empty required args", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		conv, err := svc.CreateConversation(DummyText, "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			name           string
			conversationID string
			messageText    string
			senderID       string
		}{
			{
				name:           "empty conversation ID",
				conversationID: "",
				messageText:    "Hello",
				senderID:       "user#1",
			},
			{
				name:           "empty message text",
				conversationID: conv.ID,
				messageText:    "",
				senderID:       "user#1",
			},
			{
				name:           "empty sender ID",
				conversationID: conv.ID,
				messageText:    "Hello",
				senderID:       "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Panics(t, func() {
					_, _ = svc.SendMessage(tt.conversationID, tt.messageText, tt.senderID)
				})
			})
		}
	})
}

func TestConversationService_ListMessages(t *testing.T) {
	t.Run("returns all messages from the given conversation", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)
		conv, err := svc.CreateConversation(DummyText, "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}
		msg, err := svc.SendMessage(conv.ID, "Second message", "user#1")
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
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = svc.ListMessages("")
		})
	})
	t.Run("returns error when conversation doesn't exist", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		_ = db.Open(DummyUser)
		svc := NewConversationService(db)

		// Act
		messages, err := svc.ListMessages("conv#1")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, messages)
	})
	t.Run("returns error when database query fails", func(t *testing.T) {
		// Arrange
		db := &database.Stub{QueryErr: errors.New("query err")}
		svc := NewConversationService(db)
		conv, err := svc.CreateConversation(DummyText, "user#1", "user#2")
		if err != nil {
			t.Fatal(err)
		}

		// Act
		got, err := svc.ListMessages(conv.ID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, got)
	})
}
