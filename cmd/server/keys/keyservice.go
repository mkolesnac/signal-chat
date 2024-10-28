package keys

import (
	"errors"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"math/rand"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/storage"
)

var ErrAccountNotFound = errors.New("account with the given ID not found")

type KeyService struct {
	storage storage.Provider
}

func NewKeyService(storage storage.Provider) KeyService {
	return KeyService{storage}
}

func (km *KeyService) GetPreKeyCount(accountID string) (int, error) {
	pk := models.PreKeyPartitionKey(accountID)
	skPrefix := models.PreKeySortKey("")
	var items []models.PreKey
	err := km.storage.QueryItems(pk, skPrefix, &items)
	if err != nil {
		return 0, fmt.Errorf("error getting pre key count: %v", err)
	}
	return len(items), nil
}

func (km *KeyService) GetPublicKeys(accountID string) (*GetPublicKeyResponse, error) {
	// Retrieve account model
	accPk := models.AccountPartitionKey(accountID)
	var acc models.Account
	err := km.storage.GetItem(accPk, models.AccountSortKey(accountID), &acc)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrAccountNotFound
		} else {
			return nil, fmt.Errorf("error getting account: %v", err)
		}
	}
	// Retrieve identity key
	var identityKey models.IdentityKey
	err = km.storage.GetItem(accPk, models.IdentityKeySortKey(), &identityKey)
	if err != nil {
		return nil, fmt.Errorf("error getting identity key: %v", err)
	}

	// Retrieve signed pre key
	var signedPreKey models.SignedPreKey
	err = km.storage.GetItem(accPk, models.SignedKeySortKey(acc.SignedPreKeyID), &signedPreKey)
	if err != nil {
		return nil, fmt.Errorf("error getting signed pre key: %v", err)
	}

	// Filter out pre key items
	var preKeys []models.PreKey
	err = km.storage.QueryItems(accPk, models.PreKeySortKey(""), &preKeys)
	if err != nil {
		return nil, fmt.Errorf("error getting pre keys: %v", err)
	}

	response := &GetPublicKeyResponse{
		IdentityPublicKey: identityKey.PublicKey,
		SignedPreKey: &SignedPreKeyResponse{
			KeyID:     signedPreKey.ID,
			PublicKey: signedPreKey.PublicKey,
			Signature: signedPreKey.Signature,
		},
	}

	// If there are no prekeys avaiable return response without PreKey
	count := len(preKeys)
	if count == 0 {
		return response, nil
	}

	// Randomly pick one of the pre keys and add it to the response
	index := rand.Intn(count)
	preKey := preKeys[index]
	response.PreKey = &PreKeyResponse{
		KeyID:     preKey.ID,
		PublicKey: preKey.PublicKey,
	}

	// Remove selected prekey from the storage
	err = km.storage.DeleteItem(preKey.PartitionKey, preKey.SortKey)
	if err != nil {
		return nil, fmt.Errorf("failed to delete prekey: %w", err)
	}

	return response, nil
}

func (km *KeyService) UploadNewPreKeys(accountID string, req UploadPreKeysRequest) error {
	signedPreKey := models.NewSignedPreKey(accountID, req.SignedPreKey.KeyID, [32]byte(req.SignedPreKey.PublicKey), [64]byte(req.SignedPreKey.Signature))
	err := km.storage.WriteItem(signedPreKey)
	if err != nil {
		return fmt.Errorf("failed to write signed pre key: %w", err)
	}

	var preKeys []storage.WriteableItem
	for _, p := range req.PreKeys {
		preKey := models.NewPreKey(accountID, p.KeyID, [32]byte(p.PublicKey))
		preKeys = append(preKeys, preKey)
	}
	err = km.storage.BatchWriteItems(preKeys)
	if err != nil {
		return fmt.Errorf("failed to batch write pre keys: %w", err)
	}

	return nil
}

func (km *KeyService) VerifySignature(accountID string, signedPublicKey, signature []byte) (bool, error) {
	pk := models.IdentityKeyPartitionKey(accountID)
	sk := models.IdentityKeySortKey()
	var identityKey models.IdentityKey
	err := km.storage.GetItem(pk, sk, &identityKey)
	if err != nil {
		return false, fmt.Errorf("failed to get identity key: %w", err)
	}

	result := ecc.VerifySignature(ecc.NewDjbECPublicKey(identityKey.PublicKey), signedPublicKey, [64]byte(signature))
	return result, nil
}
