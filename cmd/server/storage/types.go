package storage

import "strings"

type QueryCondition string

const (
	QUERY_BEGINS_WITH  QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix)"
	QUERY_GREATER_THAN QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix) AND sk > :sk"
	QUERY_LOWER_THAN   QueryCondition = "pk = :pk AND begins_with(sk, :skPrefix) AND sk < :sk"
)

var (
	IndexBySenderId string = "bySenderId"
)

type TableItem struct {
	PartitionKey string `dynamodbav:"pk"`
	SortKey      string `dynamodbav:"sk"`
	CreatedAt    string `dynamodbav:"createdAt"`
}

func (t TableItem) GetID() string {
	return strings.Split(t.SortKey, "#")[1]
}

type PrimaryKeyProvider interface {
	GetPartitionKey() string
	GetSortKey() string
}

type Backend interface {
	GetItem(pk, sk string, outPtr any) error
	QueryItems(pk, skPrefix string, queryCondition QueryCondition, outSlicePtr any) error
	QueryItemsBySenderID(senderID, skPrefix string, queryCondition QueryCondition, outSlicePtr any) error
	DeleteItem(pk, sk string) error
	WriteItem(item PrimaryKeyProvider) error
	BatchWriteItems(items []PrimaryKeyProvider) error
}
