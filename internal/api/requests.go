package api

type CreateAccountRequest struct {
	IdentityPublicKey []byte              `json:"identityKey" validate:"required,base64_32bytes"`
	SignedPreKey      SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
}

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKeyRequest     `json:"preKeys" validate:"required"`
}

type SendMessageRequest struct {
	CipherText string `json:"ciphertext"`
}

type SignedPreKeyRequest struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,base64_32bytes"`
	Signature []byte `json:"signature" validate:"required,base64_64bytes"`
}

type PreKeyRequest struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,base64_32bytes"`
}
