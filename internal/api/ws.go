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
	MessageTypeSyncAck         WSMessageType = "sync_ack"
	MessageTypeMessageAck      WSMessageType = "message_ack"
	MessageTypeConversationAck WSMessageType = "conversation_ack"
	MessageTypeParticipantAck  WSMessageType = "participant_ack"
)

type WSMessage struct {
	Type WSMessageType   `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WSSyncData struct {
	NewConversations []WSNewConversationPayload `json:"newConversations"`
	NewMessages      []WSNewMessagePayload      `json:"newMessages"`
}
