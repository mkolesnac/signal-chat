package services

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketManager interface {
	RegisterClient(id string, conn *websocket.Conn)
	UnregisterClient(id string)
	SendToClient(receiverID string, msg interface{}) error
}

// AccountService structure to manage client connections
type websocketManager struct {
	clients map[string]*websocket.Conn // Map of Account ID to websockets connection
	mu      sync.Mutex                 // Mutex for safe concurrent access
}

// Create a new accountService
func NewWebsocketManager() WebsocketManager {
	return &websocketManager{
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *websocketManager) RegisterClient(id string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[id] = conn
}

func (h *websocketManager) UnregisterClient(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.clients[id]; ok {
		err := conn.Close()
		if err != nil {
			// TODO: Log error
			return
		}
		delete(h.clients, id)
	}
}

// Send a message to a specific client based on the receiver ID
func (h *websocketManager) SendToClient(receiverID string, msg interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Find the client's connection by receiver ID
	if conn, ok := h.clients[receiverID]; ok {
		err := conn.WriteJSON(msg)
		if err != nil {
			return fmt.Errorf("failed to send message to client %s: %w", receiverID, err)
		}
	}
	return fmt.Errorf("client with ID %s not found", receiverID)
}
