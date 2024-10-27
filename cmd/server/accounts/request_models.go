package accounts

type CreateAccountRequest struct {
	IdentityPublicKey []byte              `json:"identityKey" validate:"required,base64_32bytes"`
	SignedPreKey      SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
}

type SignedPreKeyRequest struct {
	KeyID     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,base64_32bytes"`
	Signature []byte `json:"signature" validate:"required,base64_64bytes"`
}
