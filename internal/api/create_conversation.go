package api

type CreateConversationRequest struct {
	ConversationID string      `json:"conversationID" validate:"required,max=255"`
	Recipients     []Recipient `json:"recipients" validate:"required,min=1"`
}

type Recipient struct {
	ID                     string `json:"id" validate:"required"`
	KeyDistributionMessage []byte `json:"keyDistributionMessage,omitempty" validate:"required"`
}

type CreateConversationResponse struct {
	Error string `json:"error,omitempty"`
}

type WSNewConversationPayload struct {
	ConversationID         string   `json:"conversationID" validate:"required"`
	SenderID               string   `json:"senderId" validate:"required,max=255"`
	RecipientIDs           []string `json:"recipientIDs" validate:"required,min=1"`
	KeyDistributionMessage []byte   `json:"keyDistributionMessage" validate:"required"`
}
