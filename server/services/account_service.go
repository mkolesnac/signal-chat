package services

import (
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"signal-chat-server/models"
	"signal-chat-server/storage"
)

type AccountService interface {
	CreateAccount(name, pwd string, identityKey [32]byte, signedPrekey models.SignedPreKey, preKeys []models.PreKey) (models.Account, error)
	GetAccount(id string) (models.Account, error)
	GetSession(acc models.Account) (models.Session, error)
	GetKeyBundle(id string) (models.KeyBundle, error)
	GetPreKeyCount(acc models.Account) (int, error)
	UploadNewPreKeys(acc models.Account, signedPrekey models.SignedPreKey, preKeys []models.PreKey) error
}

type accountService struct {
	storage storage.Store
}

func NewAccountService(storage storage.Store) AccountService {
	return &accountService{storage: storage}
}

func (s *accountService) CreateAccount(name, pwd string, identityKey [32]byte, signedPreKey models.SignedPreKey, preKeys []models.PreKey) (models.Account, error) {
	signedKeyValid := ecc.VerifySignature(ecc.NewDjbECPublicKey(identityKey), signedPreKey.PublicKey, [64]byte(signedPreKey.Signature))
	if !signedKeyValid {
		return models.Account{}, ErrInvalidSignature
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return models.Account{}, fmt.Errorf("error hashing password: %w", err)
	}

	accKey := models.NewAccountPrimaryKey()
	accID := models.ToAccountID(accKey)
	timestamp := storage.GetTimestamp()

	resources := []storage.Resource{
		// Add Account profile resource
		{
			PrimaryKey:     accKey,
			CreatedAt:      timestamp,
			UpdatedAt:      timestamp,
			Name:           &name,
			PasswordHash:   pwdHash,
			SignedPreKeyID: &signedPreKey.KeyID,
		},
		// Add identity key
		{
			PrimaryKey: models.IdentityKeyPrimaryKey(accID),
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			PublicKey:  identityKey,
		},
		// Add signed pre key
		{
			PrimaryKey: models.SignedPreKeyPrimaryKey(accID, signedPreKey.KeyID),
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			PublicKey:  [32]byte(signedPreKey.PublicKey),
			Signature:  [64]byte(signedPreKey.Signature),
		},
	}
	// Add one-time prekeys
	for _, p := range preKeys {
		preKey := storage.Resource{
			PrimaryKey: models.PreKeyPrimaryKey(accID, p.KeyID),
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			PublicKey:  [32]byte(p.PublicKey),
		}
		resources = append(resources, preKey)
	}

	err = s.storage.BatchWriteItems(resources)
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to write items to storage: %w", err)
	}

	a := resources[0]
	return models.Account{
		ID:             models.ToAccountID(a.PrimaryKey),
		Name:           *a.Name,
		CreatedAt:      a.CreatedAt,
		SignedPreKeyID: *a.SignedPreKeyID,
	}, nil
}

func (s *accountService) GetAccount(id string) (models.Account, error) {
	primKey := models.AccountPrimaryKey(id)
	r, err := s.storage.GetItem(primKey.PartitionKey, primKey.SortKey)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return models.Account{}, ErrAccountNotFound
		} else {
			return models.Account{}, fmt.Errorf("error getting Account profile: %w", err)
		}
	}

	return models.Account{
		ID:             models.ToAccountID(r.PrimaryKey),
		Name:           *r.Name,
		CreatedAt:      r.CreatedAt,
		SignedPreKeyID: *r.SignedPreKeyID,
		PasswordHash:   r.PasswordHash,
	}, nil
}

func (s *accountService) GetSession(acc models.Account) (models.Session, error) {
	accKey := models.AccountPrimaryKey(acc.ID)
	items, err := s.storage.QueryItems(accKey.PartitionKey, "", storage.QueryBeginsWith)
	if err != nil {
		return models.Session{}, fmt.Errorf("error querying Account: %w", err)
	}

	session := models.Session{}
	for _, item := range items {
		if models.IsAccount(item) {
			session.Account = models.Account{
				ID:             models.ToAccountID(item.PrimaryKey),
				Name:           *item.Name,
				CreatedAt:      item.CreatedAt,
				SignedPreKeyID: *item.SignedPreKeyID,
			}
		} else if models.IsConversationMeta(item) {
			session.Conversations = append(session.Conversations, models.ConversationMeta{
				ID:                   models.ToConversationID(item.PrimaryKey),
				LastMessageSnippet:   *item.LastMessageSnippet,
				LastMessageTimestamp: *item.LastMessageTimestamp,
				LastMessageSenderID:  *item.SenderID,
			})
		}
	}

	return session, nil
}

func (s *accountService) GetKeyBundle(id string) (models.KeyBundle, error) {
	acc, err := s.GetAccount(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return models.KeyBundle{}, ErrAccountNotFound
		}
		return models.KeyBundle{}, err
	}

	// Retrieve all keys from the storage
	primKey := models.PreKeyPrimaryKey(acc.ID, "")
	items, err := s.storage.QueryItems(primKey.PartitionKey, "", storage.QueryBeginsWith)
	if err != nil {
		return models.KeyBundle{}, fmt.Errorf("error querying Account keys: %w", err)
	}

	result := models.KeyBundle{}
	// Pick only the most recent signed prekey
	signedPreKeyPrimKey := models.SignedPreKeyPrimaryKey(acc.ID, acc.SignedPreKeyID)

	var preKeyItems []storage.Resource
	for _, item := range items {
		if models.IsIdentityKey(item) {
			result.IdentityKey = item.PublicKey[:]
		} else if models.IsSignedPreKey(item) && item.SortKey == signedPreKeyPrimKey.SortKey {
			result.SignedPreKey = models.PreKey{
				KeyID:     models.ToSignedPreKeyID(item.PrimaryKey),
				PublicKey: item.PublicKey[:],
			}
		} else if models.IsPreKey(item) {
			preKeyItems = append(preKeyItems, item)
		}
	}

	if len(preKeyItems) == 0 {
		// Early exit if there are no one-time prekeys in the storage
		return result, nil
	}

	// Randomly pick one of the one-time prekeys
	index := rand.Intn(len(preKeyItems))
	preKeyRes := preKeyItems[index]
	result.PreKey = models.PreKey{
		KeyID:     models.ToPreKeyID(preKeyRes.PrimaryKey),
		PublicKey: preKeyRes.PublicKey[:],
	}

	// Delete the prekey from storage
	err = s.storage.DeleteItem(preKeyRes.PartitionKey, preKeyRes.SortKey)
	if err != nil {
		return models.KeyBundle{}, fmt.Errorf("failed to delete prekey: %w", err)
	}

	return result, nil
}

func (s *accountService) GetPreKeyCount(acc models.Account) (int, error) {
	primKey := models.PreKeyPrimaryKey(acc.ID, "")
	items, err := s.storage.QueryItems(primKey.PartitionKey, "", storage.QueryBeginsWith)
	if err != nil {
		return 0, fmt.Errorf("error getting pre key count: %w", err)
	}

	count := 0
	for _, item := range items {
		if models.IsPreKey(item) {
			count++
		}
	}
	return count, nil
}

func (s *accountService) UploadNewPreKeys(acc models.Account, signedPreKey models.SignedPreKey, preKeys []models.PreKey) error {
	// Verify signature of the signed prekey
	identityPrimaryKey := models.IdentityKeyPrimaryKey(acc.ID)
	identityKey, err := s.storage.GetItem(identityPrimaryKey.PartitionKey, identityPrimaryKey.SortKey)
	if err != nil {
		return fmt.Errorf("failed to get identity key: %w", err)
	}

	signedKeyValid := ecc.VerifySignature(ecc.NewDjbECPublicKey(identityKey.PublicKey), signedPreKey.PublicKey, [64]byte(signedPreKey.Signature))
	if !signedKeyValid {
		return ErrInvalidSignature
	}

	timestamp := storage.GetTimestamp()
	keys := []storage.Resource{
		// Add signed prekey
		{
			PrimaryKey: models.SignedPreKeyPrimaryKey(acc.ID, signedPreKey.KeyID),
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			PublicKey:  [32]byte(signedPreKey.PublicKey),
			Signature:  [64]byte(signedPreKey.Signature),
		},
	}

	// Add one-time prekeys
	for _, p := range preKeys {
		preKey := storage.Resource{
			PrimaryKey: models.PreKeyPrimaryKey(acc.ID, p.KeyID),
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			PublicKey:  [32]byte(p.PublicKey),
		}
		keys = append(keys, preKey)
	}
	err = s.storage.BatchWriteItems(keys)
	if err != nil {
		return fmt.Errorf("failed to write prekeys to storage: %w", err)
	}

	// Update signed prekey ID in Account profile
	updates := map[string]interface{}{
		"UpdatedAt":      storage.GetTimestamp(),
		"SignedPreKeyID": signedPreKey.KeyID,
	}
	accKey := acc.PrimaryKey()
	err = s.storage.UpdateItem(accKey.PartitionKey, accKey.SortKey, updates)
	if err != nil {
		return fmt.Errorf("failed to update Account's signed prekey ID: %w", err)
	}

	return nil
}
