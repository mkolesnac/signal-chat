package api

type PreKeyBundle struct {
	RegistrationID        uint32 `json:"registration_id"`
	IdentityKey           []byte `json:"identityKey"`
	SignedPreKey          PreKey `json:"signedPreKey"`
	SignedPreKeySignature []byte `json:"signedPreKeySignature"`
	PreKey                PreKey `json:"preKey"`
}
