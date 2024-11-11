package services

import (
	"fmt"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
	"signal-chat/internal/api"
)

type MessageService interface {
	GetMessages(accountID string, since int64, senderID string) ([]models.Message, error)
	SendMessage(accountID, recipientID string, req api.SendMessageRequest) (string, error)
}

type messageService struct {
	storage    storage.Backend
	accounts   AccountService
	websockets WebsocketManager
}

func NewMessageService(storage storage.Backend, accounts AccountService, websockets WebsocketManager) MessageService {
	return &messageService{storage, accounts, websockets}
}

func (s *messageService) GetMessages(accountID string, from int64, senderID string) ([]models.Message, error) {
	var messages []models.Message
	pk := models.MessagePartitionKey(accountID)
	sk := models.MessageSortKey(from)
	var err error
	if senderID == "" {
		err = s.storage.QueryItems(pk, sk, storage.QUERY_GREATER_THAN, &messages)
	} else {
		err = s.storage.QueryItemsBySenderID(senderID, sk, storage.QUERY_GREATER_THAN, &messages)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	return messages, nil
}

func (s *messageService) SendMessage(accountID, recipientID string, req api.SendMessageRequest) (string, error) {
	// Check if target account exists
	_, err := s.accounts.GetAccount(recipientID)
	if err != nil {
		return "", fmt.Errorf("error getting recipient account: %w", err)
	}

	msg := models.NewMessage(recipientID, accountID, req.CipherText)
	err = s.storage.WriteItem(msg)
	if err != nil {
		return "", fmt.Errorf("error writing message: %w", err)
	}

	_ = s.websockets.SendToClient(recipientID, msg)

	return msg.GetID(), nil
}
