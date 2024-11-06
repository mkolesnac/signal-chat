package models

type IdentityKey struct {
	PartitionKey string   `dynamodbav:"pk"`
	SortKey      string   `dynamodbav:"sk"`
	PublicKey    [32]byte `dynamodbav:"publicKey"`
}

func NewIdentityKey(accountId string, publicKey [32]byte) *IdentityKey {
	return &IdentityKey{
		PartitionKey: IdentityKeyPartitionKey(accountId),
		SortKey:      IdentityKeySortKey(),
		PublicKey:    publicKey,
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
