package api

type PreKeyBundle struct {
	RegistrationID uint32       `json:"registration_id"`
	IdentityKey    []byte       `json:"identityKey"`
	SignedPreKey   SignedPreKey `json:"signedPreKey"`
	PreKey         PreKey       `json:"preKey"`
}
