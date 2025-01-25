package api

type SignUpResponse struct {
	UserID string `json:"userId"`
}

type SignInResponse struct {
	UserID string `json:"userId"`
}

type CreateUserResponse struct {
	ID string `json:"id"`
}

type KeyBundle struct {
	IdentityKey   string `json:"identity_key"`   // Base64-encoded public key
	SignedPreKey  PreKey `json:"signed_pre_key"` // Signed pre-key details
	OneTimePreKey PreKey `json:"pre_key"`        // One-time pre-keys
}
