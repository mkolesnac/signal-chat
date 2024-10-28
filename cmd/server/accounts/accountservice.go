package accounts

import (
	"fmt"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
)

type AccountService struct {
	provider storage.Provider
}

func NewAccountService(provider storage.Provider) AccountService {
	return AccountService{provider: provider}
}

func (s *AccountService) CreateAccount(id, pwd string, req CreateAccountRequest) error {
	acc, err := models.NewAccount(id, pwd, req.SignedPreKey.KeyID)
	if err != nil {
		return err
	}
	err = s.provider.WriteItem(acc)
	if err != nil {
		return fmt.Errorf("failed to write account: %w", err)
	}

	identityKey := models.NewIdentityKey(id, [32]byte(req.IdentityPublicKey))
	err = s.provider.WriteItem(identityKey)
	if err != nil {
		return fmt.Errorf("failed to write identity key: %w", err)
	}

	signedPreKey := models.NewSignedPreKey(id, req.SignedPreKey.KeyID, [32]byte(req.SignedPreKey.PublicKey), [64]byte(req.SignedPreKey.Signature))
	err = s.provider.WriteItem(signedPreKey)
	if err != nil {
		return fmt.Errorf("failed to write signed pre key: %w", err)
	}

	return nil
}
