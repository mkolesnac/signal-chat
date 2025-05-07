package apitypes

import (
	"encoding/json"
)

type WSMessageType int

const (
	MessageTypeSync WSMessageType = iota
	MessageTypeNewMessage
	MessageTypeNewConversation
	MessageTypeParticipantAdded
	MessageTypeAck
)

type WSMessage struct {
	ID   string          `json:"id"`
	Type WSMessageType   `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type WSSyncPayload struct {
	Messages []WSMessage `json:"messages"`
}
