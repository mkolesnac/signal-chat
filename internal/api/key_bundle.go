package api

type KeyBundle struct {
	RegistrationID uint32       `json:"registrationId"`
	IdentityKey    []byte       `json:"identityKey"`
	SignedPreKey   SignedPreKey `json:"signedPreKey"`
	PreKeys        []PreKey     `json:"preKeys"`
}

type SignedPreKey struct {
	ID        uint32 `json:"id" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required"`
	Signature []byte `json:"signature" validate:"required,64bytes"`
}

type PreKey struct {
	ID        uint32 `json:"id" validate:"required"`
	PublicKey []byte `json:"publicKey" validate:"required"`
}
