package api

import (
	"encoding/json"
)

type WSHandler func(json.RawMessage) error
type WSErrorHandler func(error)

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
	Data json.RawMessage `json:"data"`
}

type WSSyncPayload struct {
	Messages []WSMessage `json:"messages"`
}
