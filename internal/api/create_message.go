package api

type CreateMessageRequest struct {
	ConversationID string `json:"conversationID" validate:"required"`
	Ciphertext     []byte `json:"ciphertext"`
}

type CreateMessageResponse struct {
	Error     string `json:"error,omitempty"`
	MessageID string `json:"messageID,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type WSNewMessagePayload struct {
	ConversationID string `json:"conversationID"`
	MessageID      string `json:"messageID"`
	SenderID       string `json:"senderID"`
	Ciphertext     []byte `json:"ciphertext"`
	Timestamp      int64  `json:"timestamp"`
}
