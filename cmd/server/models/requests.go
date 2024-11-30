package models

type CreateAccountRequest struct {
	Name              string       `json:"name" validate:"required"`
	Password          string       `json:"password" validate:"required"`
	IdentityPublicKey []byte       `json:"identityKey" validate:"required,32bytes"`
	SignedPreKey      SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys           []PreKey     `json:"preKeys" validate:"required"`
}

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKey `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKey     `json:"preKeys" validate:"required"`
}

type SendMessageRequest struct {
	RecipientID string `json:"recipientId" validate:"required"`
	CipherText  string `json:"cipherText" validate:"required"`
}
