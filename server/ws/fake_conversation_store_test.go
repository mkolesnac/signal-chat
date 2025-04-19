package ws

import "signal-chat/server/conversation"

// FakeConversationStore implements the ConversationStore interface for testing
type FakeConversationStore struct {
	conversations map[string]*conversation.Conversation
}

// NewMockConversationRepository creates a new mock conversation repository
func NewMockConversationRepository() *FakeConversationStore {
	return &FakeConversationStore{
		conversations: make(map[string]*conversation.Conversation),
	}
}

// GetConversation retrieves a conversation by ID
func (m *FakeConversationStore) GetConversation(id string) (*conversation.Conversation, error) {
	if conv, exists := m.conversations[id]; exists {
		return conv, nil
	}
	return nil, conversation.ErrConversationNotFound
}

// AddConversation adds a conversation to the mock repository
func (m *FakeConversationStore) AddConversation(id string, conv *conversation.Conversation) {
	m.conversations[id] = conv
}
