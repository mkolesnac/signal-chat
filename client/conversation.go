package main

import (
	"encoding/json"
	"fmt"
)

type Conversation struct {
	ID                   string
	LastMessagePreview   string
	LastMessageSenderID  string
	LastMessageTimestamp string
	ParticipantIDs       []string
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
