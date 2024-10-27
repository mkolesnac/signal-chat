package keys

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKeyRequest     `json:"preKeys" validate:"required"`
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
