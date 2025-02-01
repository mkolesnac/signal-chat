package models

type KeyBundle struct {
	IdentityKey  []byte `json:"identityKey"`
	SignedPreKey PreKey `json:"signedPreKey"`
	PreKey       PreKey `json:"preKey"`
}
