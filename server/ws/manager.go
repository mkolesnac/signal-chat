package ws

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"log"
	"signal-chat/internal/api"
	"signal-chat/server/conversation"
	"sync"
	"time"
)

// ConversationStore defines the interface for conversation storage operations
type ConversationStore interface {
	GetConversation(id string) (*conversation.Conversation, error)
}

// Manager manages WebSocket connections and message distribution
type Manager struct {
	// Registered clients
	clients map[string]*Client

	// Mutex to protect concurrent access to clients map
	mu sync.RWMutex

	// Database for storing messages when clients are offline
	db *badger.DB

	// Conversation repository for querying conversation data
	conversationRepo ConversationStore
}

// NewManager creates a new WebSocket manager
func NewManager(db *badger.DB, conversationRepo ConversationStore) *Manager {
	return &Manager{
		clients:          make(map[string]*Client),
		db:               db,
		conversationRepo: conversationRepo,
	}
}

// RegisterClient registers a new WebSocket connection for a user
func (m *Manager) RegisterClient(clientID string, conn Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if client already exists and close the old connection
	if existingClient, exists := m.clients[clientID]; exists {
		existingClient.Close()
	}

	// Create a new message store for this client
	messageStore := &MessageStore{
		db:       m.db,
		clientID: clientID,
	}

	// Create a new client
	client := NewClient(clientID, conn, messageStore)
	m.clients[clientID] = client

	log.Printf("Client registered: %s", clientID)
	return nil
}

// UnregisterClient removes a client from the manager
func (m *Manager) UnregisterClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[clientID]; exists {
		client.Close()
		delete(m.clients, clientID)
		log.Printf("Client unregistered: %s", clientID)
	}
}

// BroadcastNewConversation sends a notification about a new conversation to all participants
func (m *Manager) BroadcastNewConversation(senderID string, req api.NewConversationRequest) error {
	participantIDs := make([]string, 0, len(req.OtherParticipants))
	participantIDs = append(participantIDs, senderID)
	for _, participant := range req.OtherParticipants {
		participantIDs = append(participantIDs, participant.ID)
	}

	for _, participant := range req.OtherParticipants {
		payload := api.WSNewConversationPayload{
			ConversationID:         req.ConversationID,
			SenderID:               senderID,
			ParticipantIDs:         participantIDs,
			KeyDistributionMessage: participant.KeyDistributionMessage,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		m.sendMessageToClient(participant.ID, api.MessageTypeNewConversation, payloadBytes)
	}

	return nil
}

// BroadcastNewMessage sends a notification about a new message to all participants in a conversation
func (m *Manager) BroadcastNewMessage(senderID, messageID string, req api.NewMessageRequest) error {
	// Get conversation from the repository
	conv, err := m.conversationRepo.GetConversation(req.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to get conv: %w", err)
	}

	// Filter out the sender from participants
	recipientIDs := make([]string, 0, len(conv.ParticipantIDs)-1)
	for _, p := range conv.ParticipantIDs {
		if p != senderID {
			recipientIDs = append(recipientIDs, p)
		}
	}

	// Send to all recipients
	for _, id := range recipientIDs {
		// Create recipient-specific payload with all participants except the receiver
		recipientPayload := api.WSNewMessagePayload{
			ConversationID:   req.ConversationID,
			MessageID:        messageID,
			SenderID:         senderID,
			EncryptedMessage: req.EncryptedMessage,
			CreatedAt:        time.Now().Unix(),
		}

		payloadBytes, err := json.Marshal(recipientPayload)
		if err != nil {
			return err
		}

		m.sendMessageToClient(id, api.MessageTypeNewMessage, payloadBytes)
	}

	return nil
}

// sendMessageToClient sends a message to a specific client or stores it if the client is offline
func (m *Manager) sendMessageToClient(userID string, msgType api.WSMessageType, payload []byte) {
	m.mu.RLock()
	client, exists := m.clients[userID]
	m.mu.RUnlock()

	message := &api.WSMessage{
		ID:   generateMessageID(),
		Type: msgType,
		Data: payload,
	}

	if !exists {
		// Client is offline, store the message in the database
		messageStore := &MessageStore{
			db:       m.db,
			clientID: userID,
		}

		if err := messageStore.Store([]*api.WSMessage{message}); err != nil {
			log.Printf("Failed to store message for offline client %s: %v", userID, err)
		}
		return
	}

	// Client is online, send the message directly
	if err := client.SendMessage(message); err != nil {
		log.Printf("Failed to send message to client %s: %v", userID, err)
	}
}

// CloseAll closes all client connections
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for clientID, client := range m.clients {
		client.Close()
		log.Printf("Client connection closed: %s", clientID)
	}

	// Clear the clients map
	m.clients = make(map[string]*Client)
}
