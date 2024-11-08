package models

import "signal-chat/cmd/server/storage"

type IdentityKey struct {
	storage.TableItem
	PublicKey [32]byte `dynamodbav:"publicKey"`
}

func NewIdentityKey(accountId string, publicKey [32]byte) *IdentityKey {
	return &IdentityKey{
		TableItem: storage.TableItem{
			PartitionKey: IdentityKeyPartitionKey(accountId),
			SortKey:      IdentityKeySortKey(),
			CreatedAt:    getTimestamp(),
		},
		PublicKey: publicKey,
	}
}

func (i *IdentityKey) GetPartitionKey() string {
	return i.PartitionKey
}

func (i *IdentityKey) GetSortKey() string {
	return i.SortKey
}

func IdentityKeyPartitionKey(accountId string) string {
	return AccountPartitionKey(accountId)
}

func IdentityKeySortKey() string {

	return "identityKey"
}
