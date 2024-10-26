package keys

type UploadPreKeysRequest struct {
	SignedPreKey SignedPreKeyRequest `json:"signedPreKey" validate:"required"`
	PreKeys      []PreKeyRequest     `json:"preKeys" validate:"required"`
}

type SignedPreKeyRequest struct {
	KeyId     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,base64_32bytes"`
	Signature []byte `json:"signature" validate:"required,base64_64bytes"`
}

type PreKeyRequest struct {
	KeyId     string `json:"keyId" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required,base64_32bytes"`
}
