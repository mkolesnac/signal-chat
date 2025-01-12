package api

type SignUpRequest struct {
	Email             string       `json:"email" validate:"required"`
	Password          string       `json:"password" validate:"required"`
	IdentityPublicKey []byte       `json:"identityKey" validate:"required,32bytes"`
	SignedPreKey      SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys           []PreKey     `json:"preKeys" validate:"required"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKey     `json:"preKeys" validate:"required"`
}

type SendMessageRequest struct {
	CipherText string `json:"cipherText" validate:"required"`
}

type CreateConversationRequest struct {
	ParticipantIDs []string `json:"participantIds" validate:"required"`
	CipherText     string   `json:"cipherText" validate:"required"`
}
