package storage

type BatchWriteItem struct {
	PartitionKey string
	SortKey      string
	Value        any
}

type Provider interface {
	GetItem(pk, sk string, outPtr any) error
	QueryItems(pk, skPrefix string, outSlicePtr any) error
	DeleteItem(pk, sk string) error
	WriteItem(pk, sk string, value any) error
	BatchWriteItems(items []BatchWriteItem) error
}
