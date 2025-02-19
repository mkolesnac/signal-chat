package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"signal-chat/client/database"
	"signal-chat/client/models"
	"signal-chat/internal/api"
	"time"
)

type ConversationAPI interface {
	Post(route string, payload any) (int, []byte, error)
	Subscribe(eventType api.WSMessageType, handler api.WSHandler)
}

type ConversationService struct {
	db        database.Store
	apiClient ConversationAPI
}

func NewConversationService(db database.Store, apiClient ConversationAPI) *ConversationService {
	svc := &ConversationService{db: db, apiClient: apiClient}
	svc.apiClient.Subscribe(api.MessageTypeSync, svc.sync)
	return svc
}

func (c *ConversationService) sync(data json.RawMessage) error {
	var s api.WSSyncData
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, payload := range s.NewConversations {
		conv := models.Conversation{
			ID:                   payload.ConversationID,
			LastMessagePreview:   payload.MessagePreview,
			LastMessageTimestamp: payload.Timestamp,
			LastMessageSenderID:  payload.SenderID,
			ParticipantIDs:       payload.ParticipantIDs,
		}
		if err := c.writeConversation(conv); err != nil {
			return err
		}

		msg := models.Message{
			ID:        payload.MessageID,
			Text:      payload.MessageText,
			SenderID:  payload.SenderID,
			Timestamp: payload.Timestamp,
		}
		if err := c.writeMessage(conv.ID, msg); err != nil {
			return err
		}
	}

	for _, payload := range s.NewMessages {
		conv, err := c.getConversation(payload.ConversationID)
		if err != nil {
			return err
		}

		conv.LastMessagePreview = payload.Preview
		conv.LastMessageSenderID = payload.SenderID
		conv.LastMessageTimestamp = payload.Timestamp
		if err := c.writeConversation(conv); err != nil {
			return err
		}

		msg := models.Message{
			ID:        payload.MessageID,
			Text:      payload.Text,
			SenderID:  payload.SenderID,
			Timestamp: payload.Timestamp,
		}
		if err := c.writeMessage(conv.ID, msg); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConversationService) ListConversations() ([]models.Conversation, error) {
	data, err := c.db.Query(database.ConversationPK(""))
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	conversations := make([]models.Conversation, 0, len(data))
	for k, v := range data {
		conv, err := models.DeserializeConversation(v)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize conversation with key %s: %w", k, err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

func (c *ConversationService) ListMessages(conversationID string) ([]models.Message, error) {
	panicIfEmpty("conversationID", conversationID)

	_, err := c.getConversation(conversationID)
	if err != nil {
		return nil, err
	}

	data, err := c.db.Query(database.MessagePK(conversationID, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	messages := make([]models.Message, 0, len(data))
	for k, v := range data {
		msg, err := models.DeserializeMessage(v)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize message with key %s: %w", k, err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (c *ConversationService) CreateConversation(messageText, recipientID string) (models.Conversation, error) {
	panicIfEmpty("messageText", messageText)
	panicIfEmpty("recipientID", recipientID)

	msgPreview := messagePreview(messageText)
	req := api.CreateConversationRequest{
		RecipientIDs:   []string{recipientID},
		MessageText:    messageText,
		MessagePreview: msgPreview,
	}
	status, body, err := c.apiClient.Post(api.EndpointConversations, req)
	if err != nil {
		return models.Conversation{}, err
	}
	if status != http.StatusOK {
		return models.Conversation{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.CreateConversationResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.Conversation{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	conv := models.Conversation{
		ID:                   resp.ConversationID,
		LastMessagePreview:   msgPreview,
		LastMessageTimestamp: resp.Timestamp,
		LastMessageSenderID:  resp.SenderID,
		ParticipantIDs:       resp.ParticipantIDs,
	}
	if err := c.writeConversation(conv); err != nil {
		return models.Conversation{}, err
	}

	msg := models.Message{
		ID:        resp.MessageID,
		Text:      messageText,
		SenderID:  resp.SenderID,
		Timestamp: resp.Timestamp,
	}
	if err := c.writeMessage(conv.ID, msg); err != nil {
		return models.Conversation{}, err
	}

	return conv, nil
}

func (c *ConversationService) SendMessage(conversationID, messageText string) (models.Message, error) {
	panicIfEmpty("conversationID", conversationID)
	panicIfEmpty("messageText", messageText)

	conv, err := c.getConversation(conversationID)
	if err != nil {
		return models.Message{}, err
	}

	req := api.CreateMessageRequest{
		ConversationID: conv.ID,
		Text:           messageText,
		Preview:        messagePreview(messageText),
	}
	status, body, err := c.apiClient.Post(api.EndpointMessages, req)
	if err != nil {
		return models.Message{}, err
	}
	if status != http.StatusOK {
		return models.Message{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.CreateMessageResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.Message{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	msg := models.Message{
		ID:        uuid.New().String(),
		Text:      messageText,
		SenderID:  resp.SenderID,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := c.writeMessage(conv.ID, msg); err != nil {
		return models.Message{}, err
	}

	conv.LastMessagePreview = messagePreview(messageText)
	conv.LastMessageTimestamp = msg.Timestamp
	conv.LastMessageSenderID = msg.SenderID
	if err := c.writeConversation(conv); err != nil {
		return models.Message{}, err
	}

	return msg, nil
}

func messagePreview(text string) string {
	l := min(len(text), 100)
	return text[0:l]
}

func (c *ConversationService) getConversation(conversationID string) (models.Conversation, error) {
	bytes, err := c.db.Read(database.ConversationPK(conversationID))
	if err != nil {
		return models.Conversation{}, fmt.Errorf("failed to read conversation: %w", err)
	}
	if bytes == nil {
		return models.Conversation{}, fmt.Errorf("conversation not found")
	}

	conv, err := models.DeserializeConversation(bytes)
	if err != nil {
		return models.Conversation{}, fmt.Errorf("failed to deserialize conversation: %w", err)
	}

	return conv, nil
}

func (c *ConversationService) writeConversation(conv models.Conversation) error {
	bytes, err := conv.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize conversation: %w", err)
	}
	err = c.db.Write(database.ConversationPK(conv.ID), bytes)
	if err != nil {
		return fmt.Errorf("failed to write conversation: %w", err)
	}
	return nil
}

func (c *ConversationService) writeMessage(conversationID string, msg models.Message) error {
	bytes, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	err = c.db.Write(database.MessagePK(conversationID, msg.ID), bytes)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}
