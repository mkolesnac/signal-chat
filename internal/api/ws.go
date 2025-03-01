package api

import (
	"encoding/json"
)

type WSHandler func(json.RawMessage) error
type WSErrorHandler func(error)

type WSMessageType string

const (
	// Server -> Client messages
	MessageTypeSync             WSMessageType = "sync"
	MessageTypeNewMessage       WSMessageType = "new_message"
	MessageTypeNewConversation  WSMessageType = "new_conversation"
	MessageTypeParticipantAdded WSMessageType = "participant_added"

	// Client -> Server acknowledgments
	MessageTypeAck WSMessageType = "ack"
)

type WSMessage struct {
	ID   string          `json:"id"`
	Type WSMessageType   `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WSSyncData struct {
	NewConversations []WSNewConversationPayload `json:"newConversations"`
	NewMessages      []WSNewMessagePayload      `json:"newMessages"`
}

type WSAcknowledgementPayload struct {
	Type      WSMessageType `json:"type"`      // The Type of the acknowledged message
	ID        string        `json:"id"`        // The ID of the acknowledged message
	Timestamp int64         `json:"timestamp"` // When the acknowledgement was sent
}
