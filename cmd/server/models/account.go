package models

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"signal-chat/cmd/server/storage"
)

type Account struct {
	storage.TableItem
	PasswordHash   []byte `dynamodbav:"passwordHash"`
	SignedPreKeyID string `dynamodbav:"signedPreKeyID"`
}

func NewAccount(id, pwd, signedPreKeyID string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		TableItem: storage.TableItem{
			PartitionKey: AccountPartitionKey(id),
			SortKey:      AccountSortKey(id),
			CreatedAt:    getTimestamp(),
		},
		PasswordHash:   hash,
		SignedPreKeyID: signedPreKeyID,
	}, nil
}

func (a *Account) GetPartitionKey() string {
	return a.PartitionKey
}

func (a *Account) GetSortKey() string {
	return a.SortKey
}

func (a *Account) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(a.PasswordHash, []byte(password))
	return err == nil
}

func AccountPartitionKey(accountId string) string {
	return fmt.Sprintf("acc#%s", accountId)
}

func AccountSortKey(accountId string) string {
	return fmt.Sprintf("acc#%s", accountId)
}
