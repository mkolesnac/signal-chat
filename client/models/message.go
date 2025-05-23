package models

import (
	"encoding/json"
	"fmt"
	"signal-chat/client/encryption"
)

type Message struct {
	ID         string
	Text       string
	SenderID   string
	Timestamp  int64
	Ciphertext []byte
	Envelope   *encryption.Envelope
}

func (c *Message) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

func DeserializeMessage(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return Message{}, fmt.Errorf("failed to deserialize message: %w", err)
	}

	return m, nil
}
