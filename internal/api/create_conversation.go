package api

type CreateConversationRequest struct {
	RecipientIDs []string `json:"recipientIDs" validate:"required,min=1"`
	Name         string   `json:"name" validate:"required,max=255"`
}

type CreateConversationResponse struct {
	Error          string   `json:"error,omitempty"`
	ConversationID string   `json:"conversationID,omitempty"`
	ParticipantIDs []string `json:"participantIDs,omitempty"`
}

type WSNewConversationPayload struct {
	ConversationID string   `json:"conversationID"`
	Name           string   `json:"name"`
	ParticipantIDs []string `json:"participantIDs"`
}
