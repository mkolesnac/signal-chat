package database

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"os"
	"path/filepath"
)

var ErrNotInitialized = errors.New("database is already initialized. Use Open() function to open a database connection")

type database interface {
	Open(userID string) error
	Close() error
	ReadValue(pk PrimaryKey) ([]byte, error)
	WriteValue(pk PrimaryKey, v []byte) error
}

type Database struct {
	db *badger.DB
}

func (u *Database) Open(userID string) error {
	if u.db != nil {
		if err := u.Close(); err != nil {
			return err
		}
	}

	path := filepath.Join(".", "data", userID, "badger.db")
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create user database directory: %w", err)
	}

	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open Database: %w", err)
	}

	u.db = db
	return nil
}

func (u *Database) Close() error {
	if u.db != nil {
		if err := u.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		u.db = nil
	}
	return nil
}

func (u *Database) ReadValue(pk PrimaryKey) ([]byte, error) {
	u.panicIfNotInitialized()

	var bytes []byte
	err := u.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(pk))
		if err != nil {
			return err
		}

		bytes, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// Return nil if the item is not found in DB
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read value from Database: %w", err)
	}

	return bytes, nil
}

func (u *Database) QueryValues(prefix string) (map[string][]byte, error) {
	u.panicIfNotInitialized()
	if len(prefix) == 0 {
		return nil, fmt.Errorf("prefix cannot be empty")
	}

	items := make(map[string][]byte)
	err := u.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		p := []byte(prefix)
		for it.Seek(p); it.ValidForPrefix(p); it.Next() {
			item := it.Item()
			key := string(item.Key())
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			items[key] = value
		}
		return nil
	})

	return items, err
}

func (u *Database) WriteValue(pk PrimaryKey, value []byte) error {
	u.panicIfNotInitialized()

	err := u.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(pk), value)
	})

	if err != nil {
		return fmt.Errorf("failed to save the value: %w", err)
	}
	return nil
}

func (u *Database) DeleteValue(pk PrimaryKey) error {
	u.panicIfNotInitialized()
	if len(pk) == 0 {
		return fmt.Errorf("pk cannot be empty")
	}

	err := u.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(pk))
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil // pk not found, item doesn't exist
		}
		return fmt.Errorf("failed to delete the pk: %w", err)
	}

	return nil
}

func (u *Database) panicIfNotInitialized() {
	if u.db == nil {
		panic(ErrNotInitialized)
	}
}
