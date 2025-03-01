package api

type SignUpResponse struct {
	UserID string `json:"userId"`
	Error  string `json:"error,omitempty"`
}

type SignInResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Error    string `json:"error,omitempty"`
}

type GetUserResponse struct {
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

type ListUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ListUsersResponse struct {
	Users []ListUser `json:"users"`
	Error string     `json:"error,omitempty"`
}

type CreateUserResponse struct {
	ID string `json:"id"`
}

type KeyBundle struct {
	IdentityKey   string `json:"identity_key"`   // Base64-encoded public key
	SignedPreKey  PreKey `json:"signed_pre_key"` // Signed pre-key details
	OneTimePreKey PreKey `json:"pre_key"`        // One-time pre-keys
}
