package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"net/http"
	"signal-chat/client/database"
	"signal-chat/client/models"
	"signal-chat/internal/api"
)

type ConversationAPI interface {
	Post(route string, payload any) (int, []byte, error)
	Subscribe(eventType api.WSMessageType, handler api.WSHandler)
}

type ConversationService struct {
	ctx       context.Context
	db        database.Store
	apiClient ConversationAPI
}

func NewConversationService(db database.Store, apiClient ConversationAPI) *ConversationService {
	svc := &ConversationService{db: db, apiClient: apiClient}
	svc.apiClient.Subscribe(api.MessageTypeSync, svc.handleSync)
	svc.apiClient.Subscribe(api.MessageTypeNewConversation, svc.handleNewConversation)
	svc.apiClient.Subscribe(api.MessageTypeNewMessage, svc.handleNewMessage)
	return svc
}

func (c *ConversationService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *ConversationService) handleSync(data json.RawMessage) error {
	var s api.WSSyncData
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, payload := range s.NewConversations {
		conv := models.Conversation{
			ID:             payload.ConversationID,
			Name:           payload.Name,
			ParticipantIDs: payload.ParticipantIDs,
		}
		if err := c.writeConversation(conv); err != nil {
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

func (c *ConversationService) handleNewConversation(data json.RawMessage) error {
	var p api.WSNewConversationPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	conv := models.Conversation{
		ID:             p.ConversationID,
		Name:           p.Name,
		ParticipantIDs: p.ParticipantIDs,
	}
	if err := c.writeConversation(conv); err != nil {
		return err
	}

	runtime.EventsEmit(c.ctx, "conversation_added", conv)

	return nil
}

func (c *ConversationService) handleNewMessage(data json.RawMessage) error {
	var p api.WSNewMessagePayload
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	conv, err := c.getConversation(p.ConversationID)
	if err != nil {
		return err
	}

	conv.LastMessagePreview = p.Preview
	conv.LastMessageSenderID = p.SenderID
	conv.LastMessageTimestamp = p.Timestamp
	if err := c.writeConversation(conv); err != nil {
		return err
	}

	runtime.EventsEmit(c.ctx, "conversation_updated", conv)

	msg := models.Message{
		ID:        p.MessageID,
		Text:      p.Text,
		SenderID:  p.SenderID,
		Timestamp: p.Timestamp,
	}
	if err := c.writeMessage(conv.ID, msg); err != nil {
		return err
	}

	runtime.EventsEmit(c.ctx, "message_added", msg)

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

func (c *ConversationService) CreateConversation(name string, recipientIDs []string) (models.Conversation, error) {
	panicIfEmpty("name", name)
	if len(recipientIDs) == 0 {
		panic("recipientIDs must not be empty")
	}

	req := api.CreateConversationRequest{
		Name:         name,
		RecipientIDs: recipientIDs,
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
		ID:             resp.ConversationID,
		Name:           name,
		ParticipantIDs: resp.ParticipantIDs,
	}
	if err := c.writeConversation(conv); err != nil {
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
		ID:        resp.MessageID,
		Text:      messageText,
		SenderID:  resp.SenderID,
		Timestamp: resp.Timestamp,
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

	if c.ctx != nil {
		runtime.EventsEmit(c.ctx, "conversation_updated", conv)
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
