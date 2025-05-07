package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"signal-chat/internal/apitypes"
)

var (
	ErrEmailExists        = errors.New("user with same email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type UserStore struct {
	db *badger.DB
}

func (r *UserStore) CreateUser(username, password string, keyBundle apitypes.KeyBundle) (apitypes.User, error) {
	userID := uuid.New().String()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apitypes.User{}, err
	}

	keyBundleJSON, err := json.Marshal(keyBundle)
	if err != nil {
		return apitypes.User{}, err
	}

	err = r.db.Update(func(txn *badger.Txn) error {
		// Check if email exists
		_, err := txn.Get(usernameItemKey(username))
		if err == nil {
			return ErrEmailExists
		}
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		// Store user by MessageID
		err = txn.Set(userItemKey(userID), []byte(username))
		if err != nil {
			return err
		}

		// Store username->MessageID mapping
		err = txn.Set(usernameItemKey(username), []byte(userID))
		if err != nil {
			return err
		}

		// Store credentials
		err = txn.Set(credItemKey(username), hashedPassword)
		if err != nil {
			return err
		}

		return txn.Set(keyBundleItemKey(userID), keyBundleJSON)
	})

	if err != nil {
		return apitypes.User{}, err
	}

	return apitypes.User{
		ID:       userID,
		Username: username,
	}, nil
}

func (r *UserStore) GetAllUsers() ([]apitypes.User, error) {
	var users []apitypes.User

	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := userItemKey("")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			userID := string(key[len(prefix):])

			err := item.Value(func(v []byte) error {
				username := string(v)
				users = append(users, apitypes.User{
					ID:       userID,
					Username: username,
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return users, err
}

func (r *UserStore) GetUserByID(id string) (apitypes.User, error) {
	user := apitypes.User{
		ID: id,
	}

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(userItemKey(id))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			user.Username = string(val)
			return nil
		})
	})

	if err != nil {
		return apitypes.User{}, err
	}

	return user, nil
}

func (r *UserStore) GetPreKeyBundle(userID string) (apitypes.PreKeyBundle, error) {
	var preKeyBundle apitypes.PreKeyBundle

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyBundleItemKey(userID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			var keyBundle apitypes.KeyBundle
			err = json.Unmarshal(val, &keyBundle)
			if err != nil {
				return err
			}

			preKeyBundle.IdentityKey = keyBundle.IdentityKey
			preKeyBundle.SignedPreKey = keyBundle.SignedPreKey
			selected, newPreKeys, err := takeRandomItem(keyBundle.PreKeys)
			if err != nil {
				return fmt.Errorf("failed to select pre key: %w", err)
			}
			preKeyBundle.PreKey = selected

			keyBundle.PreKeys = newPreKeys
			keyBundleJSON, err := json.Marshal(keyBundle)
			if err != nil {
				return fmt.Errorf("failed to marshal key bundle: %w", err)
			}
			return txn.Set(keyBundleItemKey(userID), keyBundleJSON)
		})
	})

	if err != nil {
		return apitypes.PreKeyBundle{}, err
	}

	return preKeyBundle, nil
}

func (r *UserStore) VerifyCredentials(username, password string) (apitypes.User, error) {
	user := apitypes.User{
		Username: username,
	}
	var storedHash []byte

	err := r.db.View(func(txn *badger.Txn) error {
		usernameItem, err := txn.Get(usernameItemKey(username))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrUserNotFound
			}
			return err
		}

		err = usernameItem.Value(func(val []byte) error {
			user.ID = string(val)
			return nil
		})
		if err != nil {
			return err
		}

		credItem, err := txn.Get(credItemKey(username))
		if err != nil {
			return err
		}
		return credItem.Value(func(val []byte) error {
			storedHash = val
			return nil
		})
	})

	if err != nil {
		return apitypes.User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apitypes.User{}, err
	}
	if !bytes.Equal(storedHash, hashedPassword) {
		return apitypes.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func keyBundleItemKey(userID string) []byte {
	return []byte("keys#" + userID)
}

func credItemKey(username string) []byte {
	return []byte("cred#" + username)
}

func userItemKey(userID string) []byte {
	return []byte("user#" + userID)
}

func usernameItemKey(username string) []byte {
	return []byte("username#" + username)
}

func takeRandomItem[T any](slice []T) (T, []T, error) {
	var result T

	if len(slice) == 0 {
		return result, slice, fmt.Errorf("cannot select from empty slice")
	}

	// Generate a secure random index
	maxBig := big.NewInt(int64(len(slice)))
	randomBig, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return result, slice, fmt.Errorf("failed to generate random number: %w", err)
	}

	randomIndex := int(randomBig.Int64())

	// get the selected item
	selectedItem := slice[randomIndex]

	// Remove the item from the slice
	newSlice := append(slice[:randomIndex], slice[randomIndex+1:]...)

	return selectedItem, newSlice, nil
}
