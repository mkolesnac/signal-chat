package models

import (
	"fmt"
)

type SignedPreKey struct {
	PartitionKey string   `dynamodbav:"pk"`
	SortKey      string   `dynamodbav:"sk"`
	ID           string   `dynamodbav:"id"`
	PublicKey    [32]byte `dynamodbav:"publicKey"`
	Signature    [64]byte `dynamodbav:"signature"`
}

func NewSignedPreKey(accountId, keyId string, publicKey [32]byte, signature [64]byte) *SignedPreKey {
	return &SignedPreKey{
		PartitionKey: SignedKeyPartitionKey(accountId),
		SortKey:      SignedKeySortKey(keyId),
		PublicKey:    publicKey,
		Signature:    signature,
	}
}

func (s *SignedPreKey) GetPartitionKey() string {
	return s.PartitionKey
}

func (s *SignedPreKey) GetSortKey() string {
	return s.SortKey
}

func SignedKeyPartitionKey(accountId string) string {
	return AccountPartitionKey(accountId)
}

func SignedKeySortKey(keyId string) string {
	return fmt.Sprintf("signedPreKey#%s", keyId)
}
