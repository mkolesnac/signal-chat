package api

type SignUpRequest struct {
	UserName          string       `json:"username" validate:"required"`
	Password          string       `json:"password" validate:"required"`
	IdentityPublicKey []byte       `json:"identityKey" validate:"required,32bytes"`
	SignedPreKey      SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys           []PreKey     `json:"preKeys" validate:"required"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
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
