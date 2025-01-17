package api

type CreateUserResponse struct {
	ID string `json:"id"`
}

type KeyBundle struct {
	IdentityKey   string `json:"identity_key"`   // Base64-encoded public key
	SignedPreKey  PreKey `json:"signed_pre_key"` // Signed pre-key details
	OneTimePreKey PreKey `json:"pre_key"`        // One-time pre-keys
}
