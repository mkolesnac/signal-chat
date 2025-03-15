package models

import (
	"encoding/json"
	"fmt"
)

type ConversationMode int

const (
	OneOnOne ConversationMode = iota
	Group
)

type Conversation struct {
	ID                   string
	Name                 string
	Mode                 ConversationMode
	LastMessagePreview   string
	LastMessageSenderID  string
	LastMessageTimestamp int64
	OtherParticipantIDs  []string // IDs of the other conversation participants
}

func (c *Conversation) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func DeserializeConversation(data []byte) (Conversation, error) {
	var c Conversation
	err := json.Unmarshal(data, &c)
	if err != nil {
		return Conversation{}, fmt.Errorf("failed to deserialize Conversation: %w", err)
	}

	return c, nil
}
