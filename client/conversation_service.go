package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/client/models"
	"signal-chat/internal/api"
)

type ConversationCallback func(conv models.Conversation)

type MessageCallback func(msg models.Message)

type ConversationAPI interface {
	Get(route string) (int, []byte, error)
	Post(route string, payload any) (int, []byte, error)
	Subscribe(eventType api.WSMessageType, handler api.WSHandler)
}

type Encryptor interface {
	CreateEncryptionGroup(groupID string, recipientIDs []string) (map[string][]byte, error)
	ProcessSenderKeyDistributionMessage(groupID, senderID string, encryptedMsg []byte) error
	GroupEncrypt(groupID string, plaintext []byte) (*encryption.EncryptedMessage, error)
	GroupDecrypt(groupID, senderID string, ciphertext []byte) (*encryption.DecryptedMessage, error)
}

type ConversationService struct {
	db                  database.DB
	api                 ConversationAPI
	encryptor           Encryptor
	ConversationAdded   ConversationCallback
	ConversationUpdated ConversationCallback
	MessageAdded        MessageCallback
}

func NewConversationService(db database.DB, apiClient ConversationAPI, encryptor Encryptor) *ConversationService {
	svc := &ConversationService{
		db:        db,
		api:       apiClient,
		encryptor: encryptor,
	}

	svc.api.Subscribe(api.MessageTypeSync, svc.handleSync)
	svc.api.Subscribe(api.MessageTypeNewConversation, svc.handleNewConversation)
	svc.api.Subscribe(api.MessageTypeNewMessage, svc.handleNewMessage)

	return svc
}

func (c *ConversationService) handleSync(data json.RawMessage) error {
	var s api.WSSyncData
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, payload := range s.NewConversations {
		err := c.encryptor.ProcessSenderKeyDistributionMessage(payload.ConversationID, payload.SenderID, payload.KeyDistributionMessage)
		if err != nil {
			return err
		}

		conv := models.Conversation{
			ID:           payload.ConversationID,
			RecipientIDs: payload.RecipientIDs,
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

		decrypted, err := c.encryptor.GroupDecrypt(conv.ID, payload.SenderID, payload.EncryptedMessage)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		conv.LastMessagePreview = messagePreview(string(decrypted.Plaintext))
		conv.LastMessageSenderID = payload.SenderID
		conv.LastMessageTimestamp = payload.Timestamp
		if err := c.writeConversation(conv); err != nil {
			return err
		}

		msg := models.Message{
			ID:         payload.MessageID,
			Text:       string(decrypted.Plaintext),
			SenderID:   payload.SenderID,
			Timestamp:  payload.Timestamp,
			Ciphertext: decrypted.Ciphertext,
			Envelope:   decrypted.Envelope,
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

	err := c.encryptor.ProcessSenderKeyDistributionMessage(p.ConversationID, p.SenderID, p.KeyDistributionMessage)
	if err != nil {
		return err
	}

	conv := models.Conversation{
		ID:           p.ConversationID,
		RecipientIDs: p.RecipientIDs,
	}
	if err := c.writeConversation(conv); err != nil {
		return err
	}

	if c.ConversationAdded != nil {
		c.ConversationAdded(conv)
	}

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

	decrypted, err := c.encryptor.GroupDecrypt(conv.ID, p.SenderID, p.EncryptedMessage)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	conv.LastMessagePreview = messagePreview(string(decrypted.Plaintext))
	conv.LastMessageSenderID = p.SenderID
	conv.LastMessageTimestamp = p.Timestamp
	if err := c.writeConversation(conv); err != nil {
		return err
	}

	if c.ConversationUpdated != nil {
		c.ConversationUpdated(conv)
	}

	msg := models.Message{
		ID:         p.MessageID,
		Text:       string(decrypted.Plaintext),
		SenderID:   p.SenderID,
		Timestamp:  p.Timestamp,
		Ciphertext: decrypted.Ciphertext,
		Envelope:   decrypted.Envelope,
	}
	if err := c.writeMessage(conv.ID, msg); err != nil {
		return err
	}

	if c.MessageAdded != nil {
		c.MessageAdded(msg)
	}

	return nil
}

func (c *ConversationService) ListConversations() ([]models.Conversation, error) {
	data, err := c.db.Query(conversationKey(""))
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

	data, err := c.db.Query(messageKey(conversationID, ""))
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

func (c *ConversationService) CreateConversation(recipientIDs []string) (models.Conversation, error) {
	if len(recipientIDs) == 0 {
		panic("recipientIDs must not be empty")
	}

	id := uuid.New().String()

	keyMessages, err := c.encryptor.CreateEncryptionGroup(id, recipientIDs)
	if err != nil {
		return models.Conversation{}, fmt.Errorf("failed to generate key distribution messages: %w", err)
	}

	recipients := make([]api.Recipient, len(recipientIDs))
	for i, id := range recipientIDs {
		recipients[i] = api.Recipient{
			ID:                     id,
			KeyDistributionMessage: keyMessages[id],
		}
	}
	req := api.CreateConversationRequest{
		ConversationID: id,
		Recipients:     recipients,
	}

	status, body, err := c.api.Post(api.EndpointConversations, req)
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
		ID:           id,
		RecipientIDs: recipientIDs,
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

	encrypted, err := c.encryptor.GroupEncrypt(conv.ID, []byte(messageText))
	if err != nil {
		return models.Message{}, fmt.Errorf("failed to encrypt message content: %w", err)
	}

	req := api.CreateMessageRequest{
		ConversationID:   conv.ID,
		EncryptedMessage: encrypted.Serialized,
	}
	status, body, err := c.api.Post(api.EndpointMessages, req)
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
		ID:         resp.MessageID,
		Text:       messageText,
		Timestamp:  resp.Timestamp,
		Ciphertext: encrypted.Ciphertext,
		Envelope:   encrypted.Envelope,
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

	if c.ConversationUpdated != nil {
		c.ConversationUpdated(conv)
	}

	return msg, nil
}

func messagePreview(text string) string {
	l := min(len(text), 100)
	return text[0:l]
}

func (c *ConversationService) getConversation(conversationID string) (models.Conversation, error) {
	bytes, err := c.db.Read(conversationKey(conversationID))
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
	err = c.db.Write(conversationKey(conv.ID), bytes)
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
	err = c.db.Write(messageKey(conversationID, msg.ID), bytes)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func conversationKey(conversationID string) string {
	return fmt.Sprintf("conversation#%s", conversationID)
}

func messageKey(conversationID string, messageID string) string {
	return fmt.Sprintf("message#%s:%s", conversationID, messageID)
}
