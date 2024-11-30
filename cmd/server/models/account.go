package models

import (
	"github.com/google/uuid"
	"signal-chat/cmd/server/storage"
	"strings"
)

var accountKeyPrefix = "acc#"

type Account struct {
	SignedPreKeyID string `json:"-"` // ignore in responses
	ID             string `json:"id"`
	Name           string `json:"name"`
	CreatedAt      string `json:"createdAt"`
}

func (a *Account) GetPrimaryKey() storage.PrimaryKey {
	return storage.PrimaryKey{
		PartitionKey: accountKeyPrefix + a.ID,
		SortKey:      accountKeyPrefix + a.ID,
	}
}

func NewAccountPrimaryKey() storage.PrimaryKey {
	id := uuid.New().String()
	return GetAccountPrimaryKey(id)
}

func GetAccountPrimaryKey(id string) storage.PrimaryKey {
	return storage.PrimaryKey{
		PartitionKey: accountKeyPrefix + id,
		SortKey:      accountKeyPrefix + id,
	}
}

func IsAccount(r storage.Resource) bool {
	return strings.HasPrefix(r.SortKey, accountKeyPrefix)
}

func ToAccountID(primaryKey storage.PrimaryKey) string {
	return strings.Split(primaryKey.PartitionKey, "#")[0]
}
