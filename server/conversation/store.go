package conversation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

var (
	ErrConversationExists       = errors.New("conversation already exists")
	ErrConversationNotFound     = errors.New("conversation not found")
	ErrConversationUnauthorized = errors.New("not authorized to access specified conversation")
)

type Store struct {
	db *badger.DB
}

func NewStore(db *badger.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateConversation(participantIDs []string) (string, error) {
	id := uuid.New().String()

	err := s.db.Update(func(txn *badger.Txn) error {
		// Check if conversation already exists
		_, err := txn.Get(conversationItemKey(id))
		if err == nil {
			return ErrConversationExists
		}
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		conv := &Conversation{
			ParticipantIDs: participantIDs,
		}
		convJSON, err := json.Marshal(conv)
		if err != nil {
			return fmt.Errorf("failed to marshall conversation: %w", err)
		}
		return txn.Set(conversationItemKey(id), convJSON)
	})

	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *Store) CreateMessage(senderID, conversationID string, content []byte) (string, error) {
	msgID := uuid.New().String()

	err := s.db.Update(func(txn *badger.Txn) error {
		convItem, err := txn.Get(conversationItemKey(conversationID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrConversationNotFound
			}
			return err
		}

		var isAuthorized bool
		err = convItem.Value(func(val []byte) error {
			var conv Conversation
			err = json.Unmarshal(val, &conv)
			if err != nil {
				return err
			}

			for _, id := range conv.ParticipantIDs {
				if id == senderID {
					isAuthorized = true
					return nil
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
		if !isAuthorized {
			return ErrConversationUnauthorized
		}

		return txn.Set(messageItem(conversationID, msgID), content)
	})

	if err != nil {
		return "", err
	}

	return msgID, nil
}

func conversationItemKey(conversationID string) []byte {
	return []byte("conv#" + conversationID)
}

func messageItem(conversationID, messageID string) []byte {
	return []byte("msg#" + conversationID + ":" + messageID)
}

// GetConversation retrieves a conversation by ID
func (s *Store) GetConversation(conversationID string) (*Conversation, error) {
	var conv Conversation

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(conversationItemKey(conversationID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrConversationNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &conv)
		})
	})

	if err != nil {
		return nil, err
	}

	return &conv, nil
}
