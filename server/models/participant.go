package models

import (
	"signal-chat-server/storage"
	"strings"
)

type Participant struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	Name      string `json:"name"`
}

func ParticipantPrimaryKey(convID, accID string) storage.PrimaryKey {
	return storage.PrimaryKey{
		PartitionKey: conversationKeyPrefix + convID,
		SortKey:      accountKeyPrefix + accID,
	}
}

func IsParticipant(r storage.Resource) bool {
	return strings.HasPrefix(r.PartitionKey, conversationKeyPrefix) && strings.HasPrefix(r.SortKey, accountKeyPrefix)
}

func ToParticipantID(primKey storage.PrimaryKey) string {
	return strings.TrimPrefix(primKey.SortKey, accountKeyPrefix)
}
