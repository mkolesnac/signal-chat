package api

type CreateConversationRequest struct {
	OtherParticipantIDs []string `json:"otherParticipantIDs" validate:"required,min=1"`
	Name                string   `json:"name" validate:"required,max=255"`
	Mode                int      `json:"mode" validate:"required"`
}

type CreateConversationResponse struct {
	Error          string `json:"error,omitempty"`
	ConversationID string `json:"conversationID,omitempty"`
}

type WSNewConversationPayload struct {
	ConversationID      string   `json:"conversationID"`
	Name                string   `json:"name"`
	OtherParticipantIDs []string `json:"otherParticipantIDs"`
}
