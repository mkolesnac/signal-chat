package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"signal-chat/internal/apitypes"
	"sync"
	"sync/atomic"
	"time"
)

type pendingACK struct {
	message *apitypes.WSMessage
	sentAt  time.Time
}

type MessageStorage interface {
	Store(messages []*apitypes.WSMessage) error
	Delete(messageIDs []string) error
	LoadAll() ([]apitypes.WSMessage, error)
}

type Client struct {
	maxMessageSize int64
	writeWait      time.Duration
	readWait       time.Duration
	pingPeriod     time.Duration
	id             string
	conn           Connection
	storage        MessageStorage
	// Buffered channel of outbound messages.
	pendingACKs         map[string]pendingACK
	mu                  sync.RWMutex
	lastSentTime        int64
	counter             uint64
	closeOnce           sync.Once
	closed              atomic.Bool
	send                chan []byte
	writeDone           chan struct{}
	disconnectedHandler func()
}

func NewClient(id string, conn Connection, storage MessageStorage) *Client {
	client := &Client{
		maxMessageSize: 512,
		writeWait:      10 * time.Second,
		readWait:       60 * time.Second,
		pingPeriod:     55 * time.Second,
		id:             id,
		conn:           conn,
		storage:        storage,
		send:           make(chan []byte, 256),
		pendingACKs:    make(map[string]pendingACK),
		writeDone:      make(chan struct{}, 1),
	}

	go client.writePump()
	go client.readPump()
	go client.syncClient()

	return client
}

func (c *Client) SetDisconnectedHandler(handler func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disconnectedHandler = handler
}

func (c *Client) SendMessage(message *apitypes.WSMessage) error {
	if c.closed.Load() {
		if err := c.storage.Store([]*apitypes.WSMessage{message}); err != nil {
			return fmt.Errorf("client connection is closed; failed to store message: %w", err)
		}
		return nil
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case c.send <- payload:
		c.mu.Lock()
		c.pendingACKs[message.ID] = pendingACK{message: message, sentAt: time.Now()}
		c.mu.Unlock()
		return nil
	default:
		// send channel is closed or full
		if err := c.storage.Store([]*apitypes.WSMessage{message}); err != nil {
			return fmt.Errorf("client send channel closed or full; failed to store message: %w", err)
		}
		return nil
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		close(c.send)
		c.storePendingMessages()
	})
}

func (c *Client) handleExpiredACKs() error {
	now := time.Now()

	expired := make([]*apitypes.WSMessage, 0, len(c.pendingACKs))
	c.mu.Lock()
	for id, pending := range c.pendingACKs {
		if now.Sub(pending.sentAt) > c.readWait {
			delete(c.pendingACKs, id)
			expired = append(expired, pending.message)
		}
	}
	c.mu.Unlock()

	for _, msg := range expired {
		if err := c.storage.Store([]*apitypes.WSMessage{msg}); err != nil {
			return fmt.Errorf("failed to store message with expired ACK: %w", err)
		}
	}

	return nil
}

func (c *Client) storePendingMessages() {
	messages := make([]*apitypes.WSMessage, 0, len(c.pendingACKs))

	c.mu.RLock()
	for _, pending := range c.pendingACKs {
		messages = append(messages, pending.message)
	}
	c.mu.RUnlock()

	if err := c.storage.Store(messages); err != nil {
		log.Printf("client %s: failed to store pending messages: %v", c.id, err)
	}

	c.mu.Lock()
	c.pendingACKs = make(map[string]pendingACK)
	c.mu.Unlock()
}

func (c *Client) handleAcknowledgement(ack apitypes.WSMessage) error {
	c.mu.Lock()
	pending, ok := c.pendingACKs[ack.ID]
	if !ok {
		c.mu.Unlock()
		return nil
	}
	delete(c.pendingACKs, ack.ID)
	c.mu.Unlock()

	if pending.message.Type == apitypes.MessageTypeSync {
		var syncPayload apitypes.WSSyncPayload
		if err := json.Unmarshal(pending.message.Data, &syncPayload); err != nil {
			return err
		}

		messageIDs := make([]string, 0, len(syncPayload.Messages))
		for _, msg := range syncPayload.Messages {
			messageIDs = append(messageIDs, msg.ID)
		}

		if err := c.storage.Delete(messageIDs); err != nil {
			return fmt.Errorf("failed to delete synced messages from db: %w", err)
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

	payload := apitypes.WSSyncPayload{
		Messages: messages,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("client %s: failed to marshal stored websocket message: %v", c.id, err)
		return
	}

	syncMsg := &apitypes.WSMessage{
		ID:   generateMessageID(),
		Type: apitypes.MessageTypeSync,
		Data: payloadJSON,
	}
	if err = c.SendMessage(syncMsg); err != nil {
		log.Printf("client %s: failed to send sync message to the client: %v", c.id, err)
		return
	}
}

func (c *Client) notifyDisconnected() {
	c.mu.RLock()
	handler := c.disconnectedHandler
	c.mu.RUnlock()

	if handler != nil {
		go handler()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.writeDone <- struct{}{}
		_ = c.conn.Close()
		c.Close()
	}()

	c.conn.SetReadLimit(c.maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.readWait))

	c.conn.SetPongHandler(func(string) error {
		if err := c.handleExpiredACKs(); err != nil {
			log.Printf("client %s: failed to persist expired ACKs: %v", c.id, err)
		}
		_ = c.conn.SetReadDeadline(time.Now().Add(c.readWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("client %s: unexpected websocket error: %v", c.id, err)
			}

			c.notifyDisconnected()
			return
		}

		var wsMsg apitypes.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("client %s: malformed websocket message: %v", c.id, err)
			continue
		}

		switch wsMsg.Type {
		case apitypes.MessageTypeAck:
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
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					log.Printf("client %s: failed to write closing websocket message: %v", c.id, err)
					_ = c.conn.Close() // make readPump fail fast on error
				}
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("client %s: failed to write websocket message: %v", c.id, err)
				_ = c.conn.Close() // make readPump fail fast on error
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("client %s: failed to send websocket ping: %v", c.id, err)
				_ = c.conn.Close() // make readPump fail fast on error
				return
			}
		case <-c.writeDone:
			return
		}
	}
}
