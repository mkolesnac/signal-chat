package models

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type Account struct {
	PartitionKey   string `dynamodbav:"pk"`
	SortKey        string `dynamodbav:"sk"`
	PasswordHash   []byte `dynamodbav:"passwordHash"`
	SignedPreKeyID string `dynamodbav:"signedPreKeyID"`
}

func NewAccount(id, pwd, signedPreKeyID string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		PartitionKey:   AccountPartitionKey(id),
		SortKey:        AccountSortKey(id),
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

func (a *Account) GetID() string {
	return strings.Replace(a.PartitionKey, "acc#", "", 1)
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
