package models

type Message struct {
	PartitionKey string
	SortKey      string
	CreatedAt    string
	ID           string
	SenderID     string
	CipherText   string
}
