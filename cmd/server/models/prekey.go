package models

import (
	"fmt"
	"signal-chat/cmd/server/storage"
)

type PreKey struct {
	storage.TableItem
	PublicKey [32]byte `dynamodbav:"publicKey"`
}

func NewPreKey(accountId, keyId string, publicKey [32]byte) *PreKey {
	return &PreKey{
		TableItem: storage.TableItem{
			PartitionKey: PreKeyPartitionKey(accountId),
			SortKey:      PreKeySortKey(keyId),
			CreatedAt:    getTimestamp(),
		},
		PublicKey: publicKey,
	}
}

func (p *PreKey) GetPartitionKey() string {
	return p.PartitionKey
}

func (p *PreKey) GetSortKey() string {
	return p.SortKey
}

func PreKeyPartitionKey(accountId string) string {
	return AccountPartitionKey(accountId)
}

func PreKeySortKey(keyId string) string {
	return fmt.Sprintf("preKey#%s", keyId)
}
