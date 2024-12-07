package models

import (
	"signal-chat-server/storage"
	"strings"
)

var conversationKeyPrefix = "conv#"

type ConversationMeta struct {
	ID                   string `json:"id"`
	LastMessageSnippet   string `json:"lastMessageSnippet"`
	LastMessageTimestamp string `json:"lastMessageTimestamp"`
	LastMessageSenderID  string `json:"lastMessageSenderId"`
}

func ConversationMetaPrimaryKey(accID, convID string) storage.PrimaryKey {
	return storage.PrimaryKey{
		PartitionKey: accountKeyPrefix + accID,
		SortKey:      conversationKeyPrefix + convID,
	}
}

func IsConversationMeta(r storage.Resource) bool {
	return strings.HasPrefix(r.PartitionKey, accountKeyPrefix) && strings.HasPrefix(r.SortKey, conversationKeyPrefix)
}

func ToConversationID(primaryKey storage.PrimaryKey) string {
	return strings.Split(primaryKey.PartitionKey, "#")[1]
}
