package api

type NewConversationRequest struct {
	ConversationID    string        `json:"conversationID" validate:"required,max=255"`
	OtherParticipants []Participant `json:"otherParticipants" validate:"required,min=1"`
}

type NewConversationResponse struct {
	ConversationID string `json:"conversationID" validate:"required,max=255"`
}

type Participant struct {
	ID                     string `json:"id" validate:"required"`
	KeyDistributionMessage []byte `json:"keyDistributionMessage,omitempty" validate:"required"`
}

type WSNewConversationPayload struct {
	ConversationID         string   `json:"conversationID" validate:"required"`
	SenderID               string   `json:"senderId" validate:"required,max=255"`
	ParticipantIDs         []string `json:"participantIDs" validate:"required,min=1"`
	KeyDistributionMessage []byte   `json:"keyDistributionMessage" validate:"required"`
}
