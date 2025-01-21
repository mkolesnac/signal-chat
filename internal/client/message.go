package client

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	ID             string
	ConversationID string
	Text           string
	SenderID       string
}

func (c *Message) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func DeserializeMessage(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return Message{}, fmt.Errorf("failed to deserialize Message: %w", err)
	}

	return m, nil
}
