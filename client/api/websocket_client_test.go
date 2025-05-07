package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"signal-chat/internal/apitypes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketClient_Connect(t *testing.T) {
	t.Run("establishes websocket connection with auth token", func(t *testing.T) {
		// Arrange
		var authHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader = r.Header.Get("Authorization")
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Keep connection open for test
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		// Act
		err := client.Connect("test-token")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Bearer test-token", authHeader)
	})

	t.Run("should handle server url with trailing slash", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		// Act
		err := client.Connect(server.URL + "/")

		// Assert
		assert.NoError(t, err)
	})

	t.Run("returns error when connection fails", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "websocket error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		// Act
		err := client.Connect("test-token")

		// Assert
		assert.Error(t, err)
	})
}

func TestWebSocketClient_Close(t *testing.T) {
	t.Run("sends close message to server", func(t *testing.T) {
		// Arrange
		closeMessageReceived := atomic.Bool{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					closeMessageReceived.Store(true)
				} else {
					assert.Fail(t, "unexpected error: %w", err)
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Act
		client.Close()

		// Assert
		assert.Eventually(t, func() bool {
			return closeMessageReceived.Load()
		}, time.Second, 10*time.Millisecond, "Close message not received by server")
		assert.True(t, client.IsClosed())
	})

	t.Run("close is idempotent", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Act
		client.Close()
		initialClosedValue := client.IsClosed()
		client.Close() // Second close

		// Assert
		assert.True(t, initialClosedValue)
		assert.True(t, client.IsClosed())
	})
}

func TestWebSocketClient_MessageHandling(t *testing.T) {
	t.Run("calls handler when message is received", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Send test message
			msg := apitypes.WSMessage{
				ID:   "test-id",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"text":"Hello, world!"}`),
			}

			err := conn.WriteJSON(msg)
			require.NoError(t, err)

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		messageReceived := make(chan json.RawMessage)
		client.SetMessageHandler(apitypes.MessageTypeNewMessage, func(payload json.RawMessage) {
			messageReceived <- payload
		})

		// Act
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Assert
		select {
		case payload := <-messageReceived:
			assert.JSONEq(t, `{"text":"Hello, world!"}`, string(payload))
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for message")
		}
	})

	t.Run("sends ACK message when receiving non-ACK message", func(t *testing.T) {
		// Arrange
		ackReceived := make(chan apitypes.WSMessage)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Send test message
			msg := apitypes.WSMessage{
				ID:   "test-id",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"text":"Hello, world!"}`),
			}

			err := conn.WriteJSON(msg)
			require.NoError(t, err)

			// Wait for ACK
			var ack apitypes.WSMessage
			err = conn.ReadJSON(&ack)
			ackReceived <- ack

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		// Act
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Assert
		select {
		case ack := <-ackReceived:
			assert.Equal(t, apitypes.MessageTypeAck, ack.Type)
			assert.Equal(t, "test-id", ack.ID)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for ACK message")
		}
	})

	t.Run("doesn't call handler for different message type", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Send test message of one type
			msg := apitypes.WSMessage{
				ID:   "test-id",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"text":"Hello, world!"}`),
			}

			err := conn.WriteJSON(msg)
			require.NoError(t, err)

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		handlerCalled := false
		client.SetMessageHandler(apitypes.MessageTypeSync, func(payload json.RawMessage) {
			handlerCalled = true
		})

		// Act
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Assert - give some time for potential handler call
		time.Sleep(100 * time.Millisecond)
		assert.False(t, handlerCalled, "Handler should not be called for different message type")
	})

	t.Run("handles multiple message types", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Send first message type
			msg1 := apitypes.WSMessage{
				ID:   "msg1",
				Type: apitypes.MessageTypeNewConversation,
				Data: json.RawMessage(`{"conversation":"test1"}`),
			}
			err := conn.WriteJSON(msg1)
			require.NoError(t, err)

			// Send second message type
			msg2 := apitypes.WSMessage{
				ID:   "msg2",
				Type: apitypes.MessageTypeNewMessage,
				Data: json.RawMessage(`{"message":"test2"}`),
			}
			err = conn.WriteJSON(msg2)
			require.NoError(t, err)

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		convReceived := make(chan bool)
		msgReceived := make(chan bool)

		client.SetMessageHandler(apitypes.MessageTypeNewConversation, func(payload json.RawMessage) {
			convReceived <- true
		})

		client.SetMessageHandler(apitypes.MessageTypeNewMessage, func(payload json.RawMessage) {
			msgReceived <- true
		})

		// Act
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Assert
		select {
		case <-convReceived:
			// NewConversation message received
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for NewConversation message")
		}

		select {
		case <-msgReceived:
			// NewMessage message received
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for NewMessage message")
		}
	})
}

func TestWebSocketClient_ConnectionStateHandling(t *testing.T) {
	t.Run("notifies about connection state changes", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)

		stateChanges := make(chan ConnectionState, 2)
		client.SetConnectionStateHandler(func(state ConnectionState) {
			stateChanges <- state
		})

		// Act
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Assert
		select {
		case state := <-stateChanges:
			assert.Equal(t, StateConnected, state)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for connection state notification")
		}
	})
}

func TestWebSocketClient_IsClosed(t *testing.T) {
	t.Run("returns correct closed state", func(t *testing.T) {
		// Arrange
		client := NewWebSocketClient("ws://example.com")

		// Assert initial state
		assert.False(t, client.IsClosed())

		// Act - close the client
		client.Close()

		// Assert closed state
		assert.True(t, client.IsClosed())
	})
}

func TestWebSocketClient_Reconnect(t *testing.T) {
	t.Run("automatically reconnects when connection is closed unexpectedly", func(t *testing.T) {
		// Arrange
		connectionCount := 0
		reconnected := make(chan struct{})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectionCount++
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// If this is the first connection, close it to trigger reconnect
			if connectionCount == 1 {
				// Wait a bit before closing to ensure the client is fully connected
				time.Sleep(50 * time.Millisecond)
				err := conn.Close()
				require.NoError(t, err)
				return
			}

			// Second connection means reconnect was successful
			close(reconnected)

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)
		client.baseReconnectDelay = 10 * time.Millisecond // Reduce reconnect delay for test to make it faster
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Act & Assert
		select {
		case <-reconnected:
			assert.Equal(t, 2, connectionCount, "WebSocketClient should have reconnected once")
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for reconnection")
		}
	})

	t.Run("notifies state handler of reconnection attempts", func(t *testing.T) {
		// Arrange
		connectionCount := 0
		stateChanges := make(chan ConnectionState, 3)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectionCount++
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// If this is the first connection, close it to trigger reconnect
			if connectionCount == 1 {
				// Wait a bit before closing to ensure the client is fully connected
				time.Sleep(50 * time.Millisecond)
				err := conn.Close()
				require.NoError(t, err)
				return
			}

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)
		client.baseReconnectDelay = 10 * time.Millisecond
		client.SetConnectionStateHandler(func(state ConnectionState) {
			stateChanges <- state
		})

		err := client.Connect("test-token")
		require.NoError(t, err)

		// Act & Assert
		// First state should be connected
		select {
		case state := <-stateChanges:
			assert.Equal(t, StateConnected, state)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for initial connected state")
		}

		// Second state should be reconnecting
		select {
		case state := <-stateChanges:
			assert.Equal(t, StateReconnecting, state)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for reconnecting state")
		}

		// Third state should be connected again
		select {
		case state := <-stateChanges:
			assert.Equal(t, StateConnected, state)
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for connected state after reconnection")
		}
	})

	t.Run("does not reconnect when closed by client", func(t *testing.T) {
		// Arrange
		connectionCount := 0

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			connectionCount++
			conn, connClose := testUpgradeToWebSocket(t, w, r)
			defer connClose()

			// Keep connection open
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}))
		defer server.Close()

		client := NewWebSocketClient(server.URL)
		err := client.Connect("test-token")
		require.NoError(t, err)

		// Act
		client.Close()

		// Wait a reasonable time to ensure no reconnection attempt is made
		time.Sleep(200 * time.Millisecond)

		// Assert
		assert.Equal(t, 1, connectionCount, "WebSocketClient should not have reconnected after clean close")
		assert.True(t, client.IsClosed(), "WebSocketClient should be marked as closed")
	})
}

func testUpgradeToWebSocket(t *testing.T, w http.ResponseWriter, r *http.Request) (*websocket.Conn, func()) {
	t.Helper()

	u := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := u.Upgrade(w, r, nil)
	require.NoError(t, err)

	return conn, func() {
		_ = conn.Close()
	}
}
