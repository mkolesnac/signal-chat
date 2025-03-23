package api

type CreateMessageRequest struct {
	ConversationID   string `json:"conversationID" validate:"required"`
	EncryptedMessage []byte `json:"encryptedMessage" validate:"required"`
}

type CreateMessageResponse struct {
	Error     string `json:"error,omitempty"`
	MessageID string `json:"messageID,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type WSNewMessagePayload struct {
	ConversationID   string `json:"conversationID"`
	MessageID        string `json:"messageID"`
	SenderID         string `json:"senderID"`
	EncryptedMessage []byte `json:"encryptedMessage"`
	Timestamp        int64  `json:"timestamp"`
}
