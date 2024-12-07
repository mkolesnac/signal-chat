package api

import (
	"signal-chat-server/models"
)

type CreateAccountRequest struct {
	Name              string              `json:"name" validate:"required"`
	Password          string              `json:"password" validate:"required"`
	IdentityPublicKey []byte              `json:"identityKey" validate:"required,32bytes"`
	SignedPreKey      models.SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys           []models.PreKey     `json:"preKeys" validate:"required"`
}

type UploadPreKeysRequest struct {
	SignedPreKey models.SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys      []models.PreKey     `json:"preKeys" validate:"required"`
}

type SendMessageRequest struct {
	CipherText string `json:"cipherText" validate:"required"`
}

type CreateConversationRequest struct {
	ParticipantIDs []string `json:"participantIds" validate:"required"`
	CipherText     string   `json:"cipherText" validate:"required"`
}
