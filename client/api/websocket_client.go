package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"signal-chat/internal/apitypes"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandler func(payload json.RawMessage)

type ConnectionStateHandler func(state ConnectionState)

type ConnectionState int

const (
	StateConnected ConnectionState = iota
	StateReconnecting
)

type WebSocketClient struct {
	maxMessageSize         int64
	baseReconnectDelay     time.Duration
	maxReconnectDelay      time.Duration
	writeWait              time.Duration
	readWait               time.Duration
	conn                   *websocket.Conn
	serverURL              string
	authToken              string
	send                   chan []byte
	mu                     sync.RWMutex
	closeOnce              sync.Once
	closed                 atomic.Bool
	reconnectMu            sync.Mutex
	handlers               map[apitypes.WSMessageType][]MessageHandler
	connectionStateHandler ConnectionStateHandler
	writeDone              chan struct{}
}

func NewWebSocketClient(serverURL string) *WebSocketClient {
	wsURL := strings.Replace(strings.Replace(serverURL, "https://", "wss://", 1), "http://", "ws://", 1) + "/ws"
	return &WebSocketClient{
		maxMessageSize:     512,
		baseReconnectDelay: 1 * time.Second,
		maxReconnectDelay:  30 * time.Second,
		writeWait:          10 * time.Second,
		readWait:           60 * time.Second,
		serverURL:          wsURL,
		send:               make(chan []byte, 256),
		handlers:           make(map[apitypes.WSMessageType][]MessageHandler),
		writeDone:          make(chan struct{}, 1),
	}
}

// IsClosed returns whether the client is closed
func (c *WebSocketClient) IsClosed() bool {
	return c.closed.Load()
}

func (c *WebSocketClient) SetConnectionStateHandler(handler ConnectionStateHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionStateHandler = handler
}

// SetMessageHandler registers a handler for a specific message type
func (c *WebSocketClient) SetMessageHandler(messageType apitypes.WSMessageType, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.handlers[messageType] = append(c.handlers[messageType], handler)
}

func (c *WebSocketClient) Connect(authToken string) error {
	if c.IsClosed() {
		panic("client has been closed")
	}

	c.authToken = authToken

	header := http.Header{"Authorization": []string{"Bearer " + authToken}}
	dialer := websocket.Dialer{HandshakeTimeout: 45 * time.Second}
	conn, _, err := dialer.Dial(c.serverURL, header)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket server: %w", err)
	}

	c.conn = conn
	go c.writePump()
	go c.readPump()
	c.notifyConnectionState(StateConnected)

	return nil
}

func (c *WebSocketClient) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		close(c.send)
	})
}

// reconnect handles automatic reconnection attempts
func (c *WebSocketClient) reconnect() {
	delay := c.baseReconnectDelay
	c.notifyConnectionState(StateReconnecting)

	for {
		if c.closed.Load() {
			return
		}

		log.Printf("attempting to reconnect...")

		if err := c.Connect(c.authToken); err != nil {
			log.Printf("reconnection attempt failed: %v", err)
			time.Sleep(delay)
			delay *= 2
			if delay > c.maxReconnectDelay {
				delay = c.maxReconnectDelay
			}
			continue
		}

		break
	}

	log.Printf("successfully reconnected to WebSocket server")
}

func (c *WebSocketClient) sendACK(messageID string) {
	ack := &apitypes.WSMessage{
		ID:   messageID,
		Type: apitypes.MessageTypeAck,
	}

	payload, err := json.Marshal(ack)
	if err != nil {
		panic(fmt.Errorf("internal error: failed to marshal ACK: %w", err))
	}

	c.send <- payload
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.writeDone <- struct{}{}
		_ = c.conn.Close()

		if !c.IsClosed() {
			c.reconnect()
		}
	}()

	c.conn.SetReadLimit(c.maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.readWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.readWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("received unexpected close error: %v", err)
			}
			break
		}

		var wsMsg apitypes.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("received malformed websocket message: %v", err)
			continue
		}

		// Send ACK for received message if it's not an ACK itself
		if wsMsg.Type != apitypes.MessageTypeAck {
			c.sendACK(wsMsg.ID)
		}

		// Notify wsHandlers
		c.mu.RLock()
		handlers := c.handlers[wsMsg.Type]
		c.mu.RUnlock()

		for _, handler := range handlers {
			go handler(wsMsg.Data)
		}
	}
}

func (c *WebSocketClient) writePump() {
	// Drain old leftover signal if any
	select {
	case <-c.writeDone:
	default:
	}

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					log.Printf("failed to write closing websocket message: %v", err)
					_ = c.conn.Close() // make readPump fail fast
				}
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("failed to write websocket message: %v", err)
				_ = c.conn.Close() // make readPump fail fast
				return
			}
		case <-c.writeDone:
			return
		}
	}
}

func (c *WebSocketClient) notifyConnectionState(state ConnectionState) {
	c.mu.RLock()
	handler := c.connectionStateHandler
	c.mu.RUnlock()

	if handler != nil {
		go handler(state)
	}
}
