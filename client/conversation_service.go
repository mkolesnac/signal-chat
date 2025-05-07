package main

import (
	"encoding/json"
	"fmt"
	"log"
	"signal-chat/client/api"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/client/models"
	"signal-chat/internal/apitypes"

	"github.com/google/uuid"
)

type ConversationCallback func(conv models.Conversation)

type MessageCallback func(msg models.Message)

type ConversationAPI interface {
	CreateConversation(id string, otherParticipants []apitypes.Participant) error
	SendMessage(conversationID string, content []byte) (apitypes.SendMessageResponse, error)
	SetWSMessageHandler(messageType apitypes.WSMessageType, handler api.MessageHandler)
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

	svc.api.SetWSMessageHandler(apitypes.MessageTypeSync, func(data json.RawMessage) {
		if err := svc.handleSync(data); err != nil {
			log.Printf("error handling sync message: %v", err)
		}
	})

	svc.api.SetWSMessageHandler(apitypes.MessageTypeNewConversation, func(data json.RawMessage) {
		if err := svc.handleNewConversation(data); err != nil {
			log.Printf("error handling new conversation message: %v", err)
		}
	})

	svc.api.SetWSMessageHandler(apitypes.MessageTypeNewMessage, func(data json.RawMessage) {
		if err := svc.handleNewMessage(data); err != nil {
			log.Printf("error handling new message: %v", err)
		}
	})

	return svc
}

func (c *ConversationService) handleSync(data json.RawMessage) error {
	var syncPayload apitypes.WSSyncPayload
	if err := json.Unmarshal(data, &syncPayload); err != nil {
		return fmt.Errorf("failed to unmarshall websocket sync payload: %w", err)
	}

	for _, message := range syncPayload.Messages {
		var err error
		switch message.Type {
		case apitypes.MessageTypeNewMessage:
			err = c.handleNewMessage(message.Data)
		case apitypes.MessageTypeNewConversation:
			err = c.handleNewConversation(message.Data)
		default:
			log.Printf("unhandled websocket message type: %d", message.Type)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ConversationService) handleNewConversation(data json.RawMessage) error {
	var p apitypes.WSNewConversationPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("failed to unmarshall websocket message payload: %w", err)
	}

	err := c.encryptor.ProcessSenderKeyDistributionMessage(p.ConversationID, p.SenderID, p.KeyDistributionMessage)
	if err != nil {
		return fmt.Errorf("failed to process key distribution message: %w", err)
	}

	conv := models.Conversation{
		ID:             p.ConversationID,
		ParticipantIDs: p.ParticipantIDs,
	}
	if err := c.writeConversation(conv); err != nil {
		return fmt.Errorf("failed to store new conversation in the database: %w", err)
	}

	if c.ConversationAdded != nil {
		c.ConversationAdded(conv)
	}

	return nil
}

func (c *ConversationService) handleNewMessage(data json.RawMessage) error {
	var payload apitypes.WSNewMessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshall websocket message payload: %w", err)
	}

	conv, err := c.getConversation(payload.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to retrieve conversation for the given message: %w", err)
	}

	decrypted, err := c.encryptor.GroupDecrypt(conv.ID, payload.SenderID, payload.Content)
	if err != nil {
		return fmt.Errorf("failed to decrypt message: %w", err)
	}

	conv.LastMessagePreview = messagePreview(string(decrypted.Plaintext))
	conv.LastMessageSenderID = payload.SenderID
	conv.LastMessageTimestamp = payload.CreatedAt
	if err := c.writeConversation(conv); err != nil {
		return fmt.Errorf("failed to update conversation in the database: %w", err)
	}

	if c.ConversationUpdated != nil {
		c.ConversationUpdated(conv)
	}

	msg := models.Message{
		ID:         payload.MessageID,
		Text:       string(decrypted.Plaintext),
		SenderID:   payload.SenderID,
		Timestamp:  payload.CreatedAt,
		Ciphertext: decrypted.Ciphertext,
		Envelope:   decrypted.Envelope,
	}
	if err := c.writeMessage(conv.ID, msg); err != nil {
		return fmt.Errorf("failed to store new message in the database: %w", err)
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

	otherParticipants := make([]apitypes.Participant, len(recipientIDs))
	for i, id := range recipientIDs {
		otherParticipants[i] = apitypes.Participant{
			ID:                     id,
			KeyDistributionMessage: keyMessages[id],
		}
	}

	if err := c.api.CreateConversation(id, otherParticipants); err != nil {
		return models.Conversation{}, fmt.Errorf("failed to create conversation: %w", err)
	}

	conv := models.Conversation{
		ID:             id,
		ParticipantIDs: recipientIDs,
	}
	if err := c.writeConversation(conv); err != nil {
		return models.Conversation{}, fmt.Errorf("failed to store conversation: %w", err)
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

	resp, err := c.api.SendMessage(conv.ID, encrypted.Serialized)
	if err != nil {
		return models.Message{}, fmt.Errorf("failed to send message: %w", err)
	}

	msg := models.Message{
		ID:         resp.MessageID,
		Text:       messageText,
		Timestamp:  resp.CreatedAt,
		Ciphertext: encrypted.Ciphertext,
		Envelope:   encrypted.Envelope,
	}
	if err := c.writeMessage(conv.ID, msg); err != nil {
		return models.Message{}, fmt.Errorf("failed to store message: %w", err)
	}

	conv.LastMessagePreview = messagePreview(messageText)
	conv.LastMessageTimestamp = msg.Timestamp
	conv.LastMessageSenderID = msg.SenderID
	if err := c.writeConversation(conv); err != nil {
		return models.Message{}, fmt.Errorf("failed to store updated conversation: %w", err)
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
		return err
	}

	if err := c.db.Write(conversationKey(conv.ID), bytes); err != nil {
		return err
	}

	return nil
}

func (c *ConversationService) writeMessage(conversationID string, msg models.Message) error {
	bytes, err := msg.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if err := c.db.Write(messageKey(conversationID, msg.ID), bytes); err != nil {
		return err
	}

	return nil
}

func conversationKey(conversationID string) string {
	return fmt.Sprintf("conversation#%s", conversationID)
}

func messageKey(conversationID string, messageID string) string {
	return fmt.Sprintf("message#%s:%s", conversationID, messageID)
}
