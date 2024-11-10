package api

type CreateAccountRequest struct {
	Name              string              `json:"name" validate:"required"`
	Password          string              `json:"password" validate:"required"`
	IdentityPublicKey []byte              `json:"identityKey" validate:"required,32bytes"`
	SignedPreKey      SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
}

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKeyRequest     `json:"preKeys" validate:"required"`
}

type SendMessageRequest struct {
	RecipientID string `json:"recipientId" validate:"required"`
	CipherText  string `json:"cipherText" validate:"required"`
}

type SignedPreKeyRequest struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,32bytes"`
	Signature []byte `json:"signature" validate:"required,64bytes"`
}

type PreKeyRequest struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,32bytes"`
}
