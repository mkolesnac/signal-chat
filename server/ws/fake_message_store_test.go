package ws

import (
	"signal-chat/internal/api"
	"sync"
)

// FakeMessageStore implements a fake message store for testing
type FakeMessageStore struct {
	messages []api.WSMessage
	mu       sync.Mutex
}

func NewFakeMessageStore() *FakeMessageStore {
	return &FakeMessageStore{
		messages: make([]api.WSMessage, 0),
	}
}

func (f *FakeMessageStore) Store(messages []*api.WSMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, msg := range messages {
		f.messages = append(f.messages, *msg)
	}
	return nil
}

func (f *FakeMessageStore) Delete(messageIDs []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	newMessages := make([]api.WSMessage, 0, len(f.messages))
	for _, msg := range f.messages {
		shouldKeep := true
		for _, id := range messageIDs {
			if msg.ID == id {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newMessages = append(newMessages, msg)
		}
	}

	f.messages = newMessages
	return nil
}

func (f *FakeMessageStore) LoadAll() ([]api.WSMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := make([]api.WSMessage, len(f.messages))
	copy(result, f.messages)
	return result, nil
}
