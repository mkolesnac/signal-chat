package storage

type Resource struct {
	PrimaryKey
	CreatedAt            string   `dynamodbav:"createdAt"`
	UpdatedAt            string   `dynamodbav:"updatedAt"`
	Name                 *string  `dynamodbav:"name"`
	SenderID             *string  `dynamodbav:"senderId"`
	CipherText           *string  `dynamodbav:"cipherText"`
	LastMessageSnippet   *string  `dynamodbav:"lastMessageSnippet"`
	LastMessageTimestamp *string  `dynamodbav:"lastMessageTimestamp"`
	PublicKey            [32]byte `dynamodbav:"publicKey"`
	Signature            [64]byte `dynamodbav:"signature"`
	PasswordHash         []byte   `dynamodbav:"passwordHash"`
	SignedPreKeyID       *string  `dynamodbav:"signedPreKeyID"`
	Participants         []string `dynamodbav:"participants"`
}
