package ws

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"signal-chat/internal/apitypes"
)

type MessageStore struct {
	db       *badger.DB
	clientID string
}

func (m *MessageStore) Store(messages []*apitypes.WSMessage) error {
	wb := m.db.NewWriteBatch()
	defer wb.Cancel()

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		err = wb.Set(m.toMessageKey(msg.ID), data)
		if err != nil {
			return err
		}
	}

	err := wb.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (m *MessageStore) Delete(messageIDs []string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	wb := m.db.NewWriteBatch()
	defer wb.Cancel()

	for _, messageID := range messageIDs {
		if err := wb.Delete(m.toMessageKey(messageID)); err != nil {
			return err
		}
	}

	return wb.Flush()
}

func (m *MessageStore) LoadAll() ([]apitypes.WSMessage, error) {
	var items [][]byte

	err := m.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := m.toMessageKey("")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(v []byte) error {
				items = append(items, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	var messages []apitypes.WSMessage
	for _, item := range items {
		var msg apitypes.WSMessage
		if err := json.Unmarshal(item, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal stored websocket message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (m *MessageStore) toMessageKey(messageID string) []byte {
	return []byte(fmt.Sprintf("ws:%s:%s", m.clientID, messageID))
}
