package models

import (
	"github.com/google/uuid"
	"signal-chat/cmd/server/storage"
	"strings"
)

var messageKeyPrefix = "msg#"

type Message struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"createdAt"`
	CipherText string `json:"cipherText"`
	SenderID   string `json:"senderId"`
}

func NewMessagePrimaryKey(convID string) storage.PrimaryKey {
	id := uuid.New().String()
	return MessagePrimaryKey(convID, id)
}

func MessagePrimaryKey(convID, msgID string) storage.PrimaryKey {
	return storage.PrimaryKey{
		PartitionKey: conversationKeyPrefix + convID,
		SortKey:      messageKeyPrefix + msgID,
	}
}

func IsMessage(r storage.Resource) bool {
	return strings.HasPrefix(r.PartitionKey, conversationKeyPrefix) && strings.HasPrefix(r.SortKey, messageKeyPrefix)
}
