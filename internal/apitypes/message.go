package apitypes

type SendMessageRequest struct {
	ConversationID string `json:"conversationID" validate:"required"`
	Content        []byte `json:"content" validate:"required"`
}

type SendMessageResponse struct {
	MessageID string `json:"messageID,omitempty"`
	CreatedAt int64  `json:"timestamp,omitempty"`
}

type WSNewMessagePayload struct {
	ConversationID string `json:"conversationID"`
	MessageID      string `json:"messageID"`
	SenderID       string `json:"senderID"`
	Content        []byte `json:"content"`
	CreatedAt      int64  `json:"createdAt"`
}
