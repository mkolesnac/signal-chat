package storage

type QueryCondition string

const (
	QueryBeginsWith  QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix)"
	QueryGreaterThan QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix) AND sk > :sk"
	QueryLowerThan   QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix) AND sk < :sk"
)

type Store interface {
	GetItem(pk, sk string) (Resource, error)
	QueryItems(pk, skPrefix string, queryCondition QueryCondition) ([]Resource, error)
	DeleteItem(pk, sk string) error
	UpdateItem(pk, sk string, updates map[string]interface{}) error
	WriteItem(resource Resource) error
	BatchWriteItems(resources []Resource) error
}
