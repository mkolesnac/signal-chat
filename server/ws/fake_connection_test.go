package ws

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// FakeWebSocketConn implements a fake websocket connection for testing
type FakeWebSocketConn struct {
	readChan     chan []byte
	writeChan    chan []byte
	closeChan    chan struct{}
	closed       bool
	readDeadline time.Time
	mu           sync.Mutex
}

// NewFakeWebSocketConn creates a new fake websocket connection
func NewFakeWebSocketConn() *FakeWebSocketConn {
	return &FakeWebSocketConn{
		readChan:     make(chan []byte, 10),
		writeChan:    make(chan []byte, 10),
		closeChan:    make(chan struct{}),
		readDeadline: time.Time{}, // Zero time means no deadline
	}
}

// ReadMessage reads a message from the connection
func (f *FakeWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	f.mu.Lock()
	deadline := f.readDeadline
	f.mu.Unlock()

	// Check if deadline has already passed
	if !deadline.IsZero() && time.Now().After(deadline) {
		return 0, nil, errors.New("read deadline exceeded")
	}

	// If there's no deadline, try to read immediately
	if deadline.IsZero() {
		select {
		case msg := <-f.readChan:
			return websocket.TextMessage, msg, nil
		case <-f.closeChan:
			return 0, nil, errors.New("connection closed")
		default:
			return 0, nil, errors.New("no message available")
		}
	}

	// If there's a deadline, wait for it
	select {
	case msg := <-f.readChan:
		return websocket.TextMessage, msg, nil
	case <-f.closeChan:
		return 0, nil, errors.New("connection closed")
	case <-time.After(time.Until(deadline)):
		return 0, nil, errors.New("read deadline exceeded")
	}
}

// WriteMessage writes a message to the connection
func (f *FakeWebSocketConn) WriteMessage(messageType int, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return errors.New("connection closed")
	}

	select {
	case f.writeChan <- data:
		return nil
	default:
		panic("write buffer full")
	}

	return nil
}

// Close closes the connection
func (f *FakeWebSocketConn) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true
	close(f.closeChan)
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
func (f *FakeWebSocketConn) SetReadDeadline(t time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.readDeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
func (f *FakeWebSocketConn) SetWriteDeadline(t time.Time) error {
	// Not implemented for tests
	return nil
}

// SetPongHandler sets the handler for pong messages
func (f *FakeWebSocketConn) SetPongHandler(handler func(string) error) {
	// Not implemented for tests
}

// SetReadLimit sets the maximum size of a message read from the peer
func (f *FakeWebSocketConn) SetReadLimit(limit int64) {
	// Not implemented for tests
}
