package api

type GetPublicKeyResponse struct {
	IdentityPublicKey [32]byte              `json:"identityPublicKey"`
	SignedPreKey      *SignedPreKeyResponse `json:"signedPreKey"`
	PreKey            *PreKeyResponse       `json:"preKey"`
}

type SignedPreKeyResponse struct {
	KeyID     string   `json:"keyId"`
	PublicKey [32]byte `json:"publicKey"`
	Signature [64]byte `json:"signature"`
}

type PreKeyResponse struct {
	KeyID     string   `json:"keyId"`
	PublicKey [32]byte `json:"publicKey"`
}

type SendMessageResponse struct {
	MessageID string `json:"messageId"`
}

type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderID"`
	CipherText string `json:"cipherText"`
	CreatedAt  string `json:"createdAt,omitempty"`
}
