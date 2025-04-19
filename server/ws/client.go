package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"signal-chat/internal/api"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type pendingACK struct {
	message *api.WSMessage
	sentAt  time.Time
}

type MessageStorage interface {
	Store(messages []*api.WSMessage) error
	Delete(messageIDs []string) error
	LoadAll() ([]api.WSMessage, error)
}

type Client struct {
	id      string
	conn    Connection
	storage MessageStorage
	// Buffered channel of outbound messages.
	send        chan []byte
	pendingACKs map[string]pendingACK

	writeWait    time.Duration
	readWait     time.Duration
	pingPeriod   time.Duration
	mu           sync.RWMutex
	lastSentTime int64
	counter      uint64
	closeOnce    sync.Once
	closed       atomic.Bool
}

func NewClient(id string, conn Connection, storage MessageStorage) *Client {
	client := &Client{
		id:          id,
		conn:        conn,
		storage:     storage,
		writeWait:   10 * time.Second, // Time allowed to write a message to the peer.
		readWait:    60 * time.Second, // Time allowed to read the message from the peer.
		pingPeriod:  55 * time.Second, // Send pings to peer with this period. Must be less than readWait.
		send:        make(chan []byte, 256),
		pendingACKs: make(map[string]pendingACK),
	}

	go client.writePump()
	go client.readPump()
	go client.syncClient()

	return client
}

func (c *Client) SendMessage(message *api.WSMessage) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if c.closed.Load() {
		err := c.storage.Store([]*api.WSMessage{message})
		if err != nil {
			return fmt.Errorf("client connection is closed; failed to store message: %w", err)
		}
		return nil
	}

	c.send <- bytes

	c.mu.Lock()
	c.pendingACKs[message.ID] = pendingACK{message: message, sentAt: time.Now()}
	c.mu.Unlock()
	return nil
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		c.closed.Store(true)
		_ = c.conn.Close()

		if err := c.storePendingMessages(); err != nil {
			log.Printf("client %s: failed to store pending messages in db: %v", c.id, err)
		}
	})
}

func (c *Client) readPump() {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(c.readWait)); err != nil {
		log.Printf("client %s: failed to set read deadline: %v", c.id, err)
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.handleExpiredACKs(); err != nil {
			log.Printf("client %s: failed to persist expired ACKs: %v", c.id, err)
		}
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readWait)); err != nil {
			log.Printf("client %s: failed to set read deadline: %v", c.id, err)
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("client %s: unexpected websocket error: %v", c.id, err)
			}

			break
		}

		var wsMsg api.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("client %s: malformed WS message: %v", c.id, err)
			continue
		}

		switch wsMsg.Type {
		case api.MessageTypeAck:
			if err := c.handleAcknowledgement(wsMsg); err != nil {
				log.Printf("client %s: failed to handle ACK for message ID %s: %v", c.id, wsMsg.ID, err)
			}
		default:
			// Ignore other message types from clients
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(c.pingPeriod)

	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
				log.Printf("client %s: failed to set write deadline: %v", c.id, err)
			}

			if !ok {
				// The server has closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("client %s: failed to close websocket connection: %v", c.id, err)
				}
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("client %s: failed to write websocket message: %v", c.id, err)
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
				log.Printf("client %s: failed to set write deadline: %v", c.id, err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("client %s: failed to send websocket ping: %v", c.id, err)
				return
			}
		}
	}
}

func (c *Client) handleExpiredACKs() error {
	now := time.Now()

	expired := make([]*api.WSMessage, 0, len(c.pendingACKs))
	c.mu.Lock()
	for id, pending := range c.pendingACKs {
		if now.Sub(pending.sentAt) > c.readWait {
			delete(c.pendingACKs, id)
			expired = append(expired, pending.message)
		}
	}
	c.mu.Unlock()

	for _, msg := range expired {
		if err := c.storage.Store([]*api.WSMessage{msg}); err != nil {
			return fmt.Errorf("failed to store message with expired ACK: %w", err)
		}
	}

	return nil
}

func (c *Client) storePendingMessages() error {
	messages := make([]*api.WSMessage, 0, len(c.pendingACKs))

	c.mu.RLock()
	for _, pending := range c.pendingACKs {
		messages = append(messages, pending.message)
	}
	c.mu.RUnlock()

	if err := c.storage.Store(messages); err != nil {
		return fmt.Errorf("failed to store pending messages: %w", err)
	}

	return nil
}

func (c *Client) handleAcknowledgement(ack api.WSMessage) error {
	c.mu.Lock()
	pending, ok := c.pendingACKs[ack.ID]
	if !ok {
		c.mu.Unlock()
		return nil
	}
	delete(c.pendingACKs, ack.ID)
	c.mu.Unlock()

	if pending.message.Type == api.MessageTypeSync {
		var syncPayload api.WSSyncPayload
		if err := json.Unmarshal(pending.message.Data, &syncPayload); err != nil {
			return err
		}

		messageIDs := make([]string, 0, len(syncPayload.Messages))
		for _, msg := range syncPayload.Messages {
			messageIDs = append(messageIDs, msg.ID)
		}

		if err := c.storage.Delete(messageIDs); err != nil {
			return fmt.Errorf("failed to delete messages from db: %w", err)
		}
	}

	return nil
}

func (c *Client) syncClient() {
	messages, err := c.storage.LoadAll()
	if err != nil {
		log.Printf("client %s: failed to load websocket messages from storage: %v", c.id, err)
		return
	}
	if len(messages) == 0 {
		return
	}

	payload := api.WSSyncPayload{
		Messages: messages,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("client %s: failed to marshal stored websocket message: %v", c.id, err)
		return
	}

	syncMsg := &api.WSMessage{
		ID:   generateMessageID(),
		Type: api.MessageTypeSync,
		Data: payloadJSON,
	}
	if err = c.SendMessage(syncMsg); err != nil {
		log.Printf("client %s: failed to send sync message to the client: %v", c.id, err)
		return
	}
}
