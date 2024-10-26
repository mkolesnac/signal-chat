package storage

type WriteableItem interface {
	GetPartitionKey() string
	GetSortKey() string
}

type Provider interface {
	GetItem(pk, sk string, outPtr any) error
	QueryItems(pk, skPrefix string, outSlicePtr any) error
	DeleteItem(pk, sk string) error
	WriteItem(item WriteableItem) error
	BatchWriteItems(items []WriteableItem) error
}
