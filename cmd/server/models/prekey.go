package models

import "fmt"

type PreKey struct {
	PartitionKey string   `dynamodbav:"pk"`
	SortKey      string   `dynamodbav:"sk"`
	ID           string   `dynamodbav:"id"`
	PublicKey    [32]byte `dynamodbav:"publicKey"`
}

func NewPreKey(accountId, keyId string, publicKey [32]byte) *PreKey {
	return &PreKey{
		PartitionKey: PreKeyPartitionKey(accountId),
		SortKey:      PreKeySortKey(keyId),
		PublicKey:    publicKey,
	}
}

func (p *PreKey) GetPartitionKey() string {
	return p.PartitionKey
}

func (p *PreKey) GetSortKey() string {
	return p.SortKey
}

func PreKeyPartitionKey(accountId string) string {
	return fmt.Sprintf("acc#%s", accountId)
}

func PreKeySortKey(keyId string) string {
	return fmt.Sprintf("signedPreKey#%s", keyId)
}
