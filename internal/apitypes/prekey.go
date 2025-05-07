package apitypes

type GetPreKeyBundleResponse struct {
	PreKeyBundle PreKeyBundle `json:"preKeyBundle"`
}

type PreKeyBundle struct {
	RegistrationID uint32       `json:"registrationId"`
	IdentityKey    []byte       `json:"identityKey"`
	SignedPreKey   SignedPreKey `json:"signedPreKey"`
	PreKey         PreKey       `json:"preKey"`
}
