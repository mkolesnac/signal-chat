package api

type CreateConversationRequest struct {
	RecipientIDs   []string `json:"recipientIDs" validate:"required,min=1"`
	MessageText    string   `json:"messageText" validate:"required"`
	MessagePreview string   `json:"messagePreview" validate:"required"`
}

type CreateConversationResponse struct {
	Error          string   `json:"error,omitempty"`
	ConversationID string   `json:"conversationID,omitempty"`
	MessageID      string   `json:"messageID,omitempty"`
	SenderID       string   `json:"senderID,omitempty"`
	ParticipantIDs []string `json:"participantIDs,omitempty"`
	Timestamp      string   `json:"timestamp,omitempty"`
}

type WSNewConversationPayload struct {
	ConversationID string   `json:"conversationID"`
	ParticipantIDs []string `json:"participantIDs"`
	SenderID       string   `json:"senderID"`
	MessageID      string   `json:"messageID"`
	MessageText    string   `json:"messageText" validate:"required"`
	MessagePreview string   `json:"messagePreview"`
	Timestamp      string   `json:"timestamp"`
}
