package api

type PreKey struct {
	ID        uint32 `json:"id" validate:"required"`
	PublicKey []byte `json:"public_key" validate:"required,32bytes"`
}

type SignedPreKey struct {
	ID        uint32 `json:"id" validate:"required"`
	PublicKey []byte `json:"public_key" validate:"required,32bytes"`
	Signature []byte `json:"signature" validate:"required,64bytes"`
}
