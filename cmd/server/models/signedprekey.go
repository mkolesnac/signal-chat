package models

import (
	"fmt"
	"signal-chat/cmd/server/storage"
)

type SignedPreKey struct {
	storage.TableItem
	PublicKey [32]byte `dynamodbav:"publicKey"`
	Signature [64]byte `dynamodbav:"signature"`
}

func NewSignedPreKey(accountId, keyId string, publicKey [32]byte, signature [64]byte) *SignedPreKey {
	return &SignedPreKey{
		TableItem: storage.TableItem{
			PartitionKey: SignedKeyPartitionKey(accountId),
			SortKey:      SignedKeySortKey(keyId),
			CreatedAt:    getTimestamp(),
		},
		PublicKey: publicKey,
		Signature: signature,
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
