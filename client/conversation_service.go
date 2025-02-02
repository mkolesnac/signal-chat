package main

import (
	"fmt"
	"github.com/google/uuid"
	"signal-chat/client/database"
	"time"
)

type ConversationService struct {
	db database.Store
}

func NewConversationService(db database.Store) *ConversationService {
	return &ConversationService{db: db}
}

func (c *ConversationService) ListConversations() ([]Conversation, error) {
	data, err := c.db.Query(database.ConversationPK(""))
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	conversations := make([]Conversation, 0, len(data))
	for k, v := range data {
		conv, err := DeserializeConversation(v)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize conversation with key %s: %w", k, err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (c *ConversationService) ListMessages(conversationID string) ([]Message, error) {
	requireNonEmpty("conversationID", conversationID)

	_, err := c.getConversation(conversationID)
	if err != nil {
		return nil, err
	}

	data, err := c.db.Query(database.MessagePK(conversationID, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	messages := make([]Message, 0, len(data))
	for k, v := range data {
		msg, err := DeserializeMessage(v)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize message with key %s: %w", k, err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (c *ConversationService) CreateConversation(messageText, senderID, recipientID string) (Conversation, error) {
	requireNonEmpty("messageText", messageText)
	requireNonEmpty("senderID", senderID)
	requireNonEmpty("recipientID", recipientID)

	conv := Conversation{
		ID:                   uuid.New().String(),
		LastMessagePreview:   messagePreview(messageText),
		LastMessageSenderID:  senderID,
		LastMessageTimestamp: time.Now().Format(time.RFC3339),
		ParticipantIDs:       []string{senderID, recipientID},
	}

	bytes, err := conv.Serialize()
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to serialize conversation: %w", err)
	}
	err = c.db.Write(database.ConversationPK(conv.ID), bytes)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to write conversation: %w", err)
	}

	_, err = c.SendMessage(conv.ID, messageText, senderID)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to create message: %w", err)
	}

	return conv, nil
}

func (c *ConversationService) SendMessage(conversationID, messageText, senderID string) (Message, error) {
	requireNonEmpty("conversationID", conversationID)
	requireNonEmpty("messageText", messageText)
	requireNonEmpty("senderID", senderID)

	conv, err := c.getConversation(conversationID)
	if err != nil {
		return Message{}, err
	}

	msg := Message{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		Text:           messageText,
		SenderID:       senderID,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	bytes, err := msg.Serialize()
	if err != nil {
		return Message{}, fmt.Errorf("failed to serialize message: %w", err)
	}

	err = c.db.Write(database.MessagePK(conversationID, msg.ID), bytes)
	if err != nil {
		return Message{}, fmt.Errorf("failed to write message: %w", err)
	}

	conv.LastMessagePreview = messagePreview(messageText)
	conv.LastMessageTimestamp = msg.Timestamp
	conv.LastMessageSenderID = msg.SenderID
	convBytes, err := conv.Serialize()
	if err != nil {
		return Message{}, fmt.Errorf("failed to serialize conversation: %w", err)
	}

	err = c.db.Write(database.ConversationPK(conversationID), convBytes)
	if err != nil {
		return Message{}, fmt.Errorf("failed to update conversation: %w", err)
	}

	return msg, nil
}

func messagePreview(text string) string {
	l := min(len(text), 100)
	return text[0:l]
}

func (c *ConversationService) getConversation(conversationID string) (Conversation, error) {
	bytes, err := c.db.Read(database.ConversationPK(conversationID))
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to read conversation: %w", err)
	}
	if bytes == nil {
		return Conversation{}, fmt.Errorf("conversation not found")
	}

	conv, err := DeserializeConversation(bytes)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to deserialize conversation: %w", err)
	}

	return conv, nil
}
