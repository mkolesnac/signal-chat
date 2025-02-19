package api

type CreateMessageRequest struct {
	ConversationID string `json:"conversationID" validate:"required"`
	Text           string `json:"text" validate:"required"`
	Preview        string `json:"preview" validate:"required"`
}

type CreateMessageResponse struct {
	Error     string `json:"error,omitempty"`
	MessageID string `json:"messageID,omitempty"`
	SenderID  string `json:"senderID,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

type WSNewMessagePayload struct {
	ConversationID string `json:"conversationID"`
	MessageID      string `json:"messageID"`
	SenderID       string `json:"senderID"`
	Text           string `json:"text"`
	Preview        string `json:"preview"`
	Timestamp      string `json:"timestamp"`
}
