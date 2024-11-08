package services

import (
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"math/rand"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
	"signal-chat/internal/api"
)

type KeyService struct {
	storage  storage.Backend
	accounts AccountService
}

func NewKeyService(storage storage.Backend, accounts AccountService) *KeyService {
	return &KeyService{storage, accounts}
}

func (s *KeyService) GetPreKeyCount(accountID string) (int, error) {
	pk := models.PreKeyPartitionKey(accountID)
	skPrefix := models.PreKeySortKey("")
	var items []models.PreKey
	err := s.storage.QueryItems(pk, skPrefix, storage.BEGINS_WITH, &items)
	if err != nil {
		return 0, fmt.Errorf("error getting pre key count: %w", err)
	}
	return len(items), nil
}

func (s *KeyService) GetPublicKeys(accountID string) (*api.GetPublicKeyResponse, error) {
	// Retrieve account model
	acc, err := s.accounts.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("error getting account: %w", err)
	}

	// Retrieve identity key
	var identityKey models.IdentityKey
	err = s.storage.GetItem(acc.PartitionKey, models.IdentityKeySortKey(), &identityKey)
	if err != nil {
		return nil, fmt.Errorf("error getting identity key: %w", err)
	}

	// Retrieve signed pre key
	var signedPreKey models.SignedPreKey
	err = s.storage.GetItem(acc.PartitionKey, models.SignedKeySortKey(acc.SignedPreKeyID), &signedPreKey)
	if err != nil {
		return nil, fmt.Errorf("error getting signed pre key: %w", err)
	}

	// Filter out pre key items
	var preKeys []models.PreKey
	err = s.storage.QueryItems(acc.PartitionKey, models.PreKeySortKey(""), storage.BEGINS_WITH, &preKeys)
	if err != nil {
		return nil, fmt.Errorf("error getting pre keys: %w", err)
	}

	response := &api.GetPublicKeyResponse{
		IdentityPublicKey: identityKey.PublicKey,
		SignedPreKey: &api.SignedPreKeyResponse{
			KeyID:     signedPreKey.GetID(),
			PublicKey: signedPreKey.PublicKey,
			Signature: signedPreKey.Signature,
		},
	}

	// If there are no prekeys available return response without PreKeyRequest
	count := len(preKeys)
	if count == 0 {
		return response, nil
	}

	// Randomly pick one of the pre keys and add it to the response
	index := rand.Intn(count)
	preKey := preKeys[index]
	response.PreKey = &api.PreKeyResponse{
		KeyID:     preKey.ID,
		PublicKey: preKey.PublicKey,
	}

	// Remove selected prekey from the storage
	err = s.storage.DeleteItem(preKey.PartitionKey, preKey.SortKey)
	if err != nil {
		return nil, fmt.Errorf("failed to delete prekey: %w", err)
	}

	return response, nil
}

func (s *KeyService) UploadNewPreKeys(accountID string, req api.UploadPreKeysRequest) error {
	// Retrieve account model
	acc, err := s.accounts.GetAccount(accountID)
	if err != nil {
		return fmt.Errorf("error getting account: %w", err)
	}

	// Update signed prekey ID
	acc.SignedPreKeyID = req.SignedPreKey.KeyID
	err = s.storage.WriteItem(acc)

	// Write new signed prekey to db
	signedPreKey := models.NewSignedPreKey(accountID, req.SignedPreKey.KeyID, [32]byte(req.SignedPreKey.PublicKey), [64]byte(req.SignedPreKey.Signature))
	err = s.storage.WriteItem(signedPreKey)
	if err != nil {
		return fmt.Errorf("failed to write signed pre key: %w", err)
	}

	// Write one-time prekeys
	var preKeys []storage.PrimaryKeyProvider
	for _, p := range req.PreKeys {
		preKey := models.NewPreKey(accountID, p.KeyID, [32]byte(p.PublicKey))
		preKeys = append(preKeys, preKey)
	}
	err = s.storage.BatchWriteItems(preKeys)
	if err != nil {
		return fmt.Errorf("failed to batch write pre keys: %w", err)
	}

	return nil
}

func (s *KeyService) VerifySignature(accountID string, signedPublicKey, signature []byte) (bool, error) {
	pk := models.IdentityKeyPartitionKey(accountID)
	sk := models.IdentityKeySortKey()
	var identityKey models.IdentityKey
	err := s.storage.GetItem(pk, sk, &identityKey)
	if err != nil {
		return false, fmt.Errorf("failed to get identity key: %w", err)
	}

	result := ecc.VerifySignature(ecc.NewDjbECPublicKey(identityKey.PublicKey), signedPublicKey, [64]byte(signature))
	return result, nil
}
