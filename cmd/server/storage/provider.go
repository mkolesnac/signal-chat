package storage

type QueryCondition string

const (
	BEGINS_WITH  QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix)"
	GREATER_THAN QueryCondition = "pk = :pk AND sk > :skPrefix"
	LOWER_THAN   QueryCondition = "pk = :pk AND sk < :skPrefix"
)

type TableItem interface {
	GetPartitionKey() string
	GetSortKey() string
}

type Provider interface {
	GetItem(pk, sk string, outPtr any) error
	QueryItems(pk, skPrefix string, queryCondition QueryCondition, outSlicePtr any) error
	DeleteItem(pk, sk string) error
	WriteItem(item TableItem) error
	BatchWriteItems(items []TableItem) error
}
