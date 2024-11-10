package services

import (
	"errors"
	"fmt"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
	"signal-chat/internal/api"
)

type AccountService interface {
	CreateAccount(name, pwd string, req api.CreateAccountRequest) (string, error)
	GetAccount(id string) (*models.Account, error)
}

type accountService struct {
	storage storage.Backend
}

func NewAccountService(storage storage.Backend) AccountService {
	return &accountService{storage: storage}
}

func (s *accountService) CreateAccount(name, pwd string, req api.CreateAccountRequest) (string, error) {
	acc, err := models.NewAccount(name, pwd, req.SignedPreKey.KeyID)
	if err != nil {
		return "", err
	}
	err = s.storage.WriteItem(acc)
	if err != nil {
		return "", fmt.Errorf("failed to write account: %w", err)
	}
	var accID = acc.GetID()

	identityKey := models.NewIdentityKey(accID, [32]byte(req.IdentityPublicKey))
	err = s.storage.WriteItem(identityKey)
	if err != nil {
		return "", fmt.Errorf("failed to write identity key: %w", err)
	}

	signedPreKey := models.NewSignedPreKey(accID, req.SignedPreKey.KeyID, [32]byte(req.SignedPreKey.PublicKey), [64]byte(req.SignedPreKey.Signature))
	err = s.storage.WriteItem(signedPreKey)
	if err != nil {
		return "", fmt.Errorf("failed to write signed pre key: %w", err)
	}

	return accID, nil
}

func (s *accountService) GetAccount(id string) (*models.Account, error) {
	accPk := models.AccountPartitionKey(id)
	var acc models.Account
	err := s.storage.GetItem(accPk, models.AccountSortKey(id), &acc)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrAccountNotFound
		} else {
			return nil, fmt.Errorf("error getting account: %w", err)
		}
	}

	return &acc, nil
}
