package services

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"signal-chat/server/models"
	"signal-chat/server/storage"
	"strings"
)

type ConversationService interface {
	CreateConversation(acc models.Account, cipherText string, participantIDs []string) (models.Message, error)
	GetConversation(acc models.Account, conversationId string) (models.Conversation, error)
	SendMessage(acc models.Account, conversationId, cipherText string) (models.Message, error)
}

type conversationService struct {
	storage    storage.Store
	websockets WebsocketManager
}

func NewConversationService(storage storage.Store, websockets WebsocketManager) ConversationService {
	return &conversationService{storage, websockets}
}

func (s *conversationService) CreateConversation(acc models.Account, cipherText string, participantIDs []string) (models.Message, error) {
	timestamp := storage.GetTimestamp()
	convId := uuid.New().String()
	msgId := uuid.New().String()

	for _, id := range participantIDs {
		if id == acc.ID {
			return models.Message{}, errors.New("account ID cannot be included in the list of participant IDs")
		}
	}

	items := []storage.Resource{
		// Conversation meta for current Account
		{
			PrimaryKey:           models.ConversationMetaPrimaryKey(acc.ID, convId),
			CreatedAt:            timestamp,
			SenderID:             &acc.ID,
			LastMessageSnippet:   &cipherText,
			LastMessageTimestamp: &timestamp,
		},
		// Message
		{
			PrimaryKey: models.MessagePrimaryKey(convId, msgId),
			CreatedAt:  timestamp,
			CipherText: &cipherText,
			SenderID:   &acc.ID,
		},
	}

	for _, id := range participantIDs {
		// Create meta for each participant
		items = append(items, storage.Resource{
			PrimaryKey:           models.ConversationMetaPrimaryKey(id, convId),
			CreatedAt:            timestamp,
			SenderID:             &acc.ID,
			LastMessageSnippet:   &cipherText,
			LastMessageTimestamp: &timestamp,
		})
	}

	err := s.storage.BatchWriteItems(items)
	if err != nil {
		return models.Message{}, fmt.Errorf("failed to write items to storage: %w", err)
	}

	msg := models.Message{
		ID:         msgId,
		CreatedAt:  timestamp,
		CipherText: cipherText,
		SenderID:   acc.ID,
	}
	return msg, nil
}

func (s *conversationService) GetConversation(acc models.Account, conversationId string) (models.Conversation, error) {
	primKey := models.MessagePrimaryKey(conversationId, "")
	items, err := s.storage.QueryItems(primKey.PartitionKey, "", storage.QueryBeginsWith)
	if err != nil {
		return models.Conversation{}, fmt.Errorf("error querying conversation: %w", err)
	}
	if len(items) == 0 {
		return models.Conversation{}, ErrConversationNotFound
	}

	// Check if Account is authorized to access the conversation
	err = checkIfParticipant(acc.ID, items)
	if err != nil {
		return models.Conversation{}, ErrUnauthorized
	}

	conv := models.Conversation{}
	for _, item := range items {
		if models.IsMessage(item) {
			m := models.Message{
				ID:         models.ToAccountID(item.PrimaryKey),
				CreatedAt:  item.CreatedAt,
				CipherText: *item.CipherText,
				SenderID:   *item.SenderID,
			}
			conv.Messages = append(conv.Messages, m)
		} else if models.IsParticipant(item) {
			p := models.Participant{
				ID:        models.ToParticipantID(item.PrimaryKey),
				CreatedAt: item.CreatedAt,
				Name:      *item.Name,
			}
			conv.Participants = append(conv.Participants, p)
		}
	}

	return conv, nil
}

func (s *conversationService) SendMessage(acc models.Account, conversationId, cipherText string) (models.Message, error) {
	// Retrieve all participants of the given conversation
	participantPrimKey := models.ParticipantPrimaryKey(conversationId, "")
	participantPrefix := strings.Split(participantPrimKey.SortKey, "#")[0] + "#"
	participants, err := s.storage.QueryItems(participantPrimKey.PartitionKey, participantPrefix, storage.QueryBeginsWith)
	if err != nil {
		return models.Message{}, fmt.Errorf("error querying conversation: %w", err)
	}
	if len(participants) == 0 {
		return models.Message{}, ErrConversationNotFound
	}

	// Check if the sender is participant in the conversation
	err = checkIfParticipant(acc.ID, participants)
	if err != nil {
		return models.Message{}, ErrUnauthorized
	}

	timestamp := storage.GetTimestamp()
	msgId := uuid.New().String()
	items := []storage.Resource{
		// Message
		{
			PrimaryKey: models.MessagePrimaryKey(conversationId, msgId),
			CreatedAt:  timestamp,
			SenderID:   &acc.ID,
			CipherText: &cipherText,
		},
	}
	// Update meta for all participants
	for _, r := range participants {
		id := models.ToParticipantID(r.PrimaryKey)
		m := storage.Resource{
			PrimaryKey:           models.ConversationMetaPrimaryKey(id, conversationId),
			CreatedAt:            timestamp,
			SenderID:             &acc.ID,
			LastMessageSnippet:   &cipherText,
			LastMessageTimestamp: &timestamp,
		}
		items = append(items, m)
	}

	err = s.storage.BatchWriteItems(items)
	if err != nil {
		return models.Message{}, fmt.Errorf("failed to write items: %w", err)
	}

	msg := models.Message{
		ID:         msgId,
		CreatedAt:  timestamp,
		CipherText: cipherText,
		SenderID:   acc.ID,
	}

	// Send notification to every conversation participant except the sender
	for _, p := range participants {
		id := models.ToParticipantID(p.PrimaryKey)
		if id == acc.ID {
			continue
		}
		_ = s.websockets.SendToClient(id, msg)
	}

	return msg, nil
}

func checkIfParticipant(accID string, items []storage.Resource) error {
	accPrimKey := models.AccountPrimaryKey(accID)
	for _, r := range items {
		if r.SortKey == accPrimKey.SortKey {
			return nil
		}
	}

	return ErrUnauthorized
}
