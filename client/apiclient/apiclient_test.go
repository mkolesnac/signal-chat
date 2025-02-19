package apiclient

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"signal-chat/internal/api"
	"testing"
	"time"
)

type TestData struct {
	Value string
}

const (
	DummyURL      = "http://localhost:8080"
	DummyRoute    = "/dummy"
	DummyUsername = "dummy"
	DummyPassword = "dummy"
)

var (
	DummyRoundTripper = &SpyRoundTripper{}
)

func TestAPIClient_StartSession(t *testing.T) {
	t.Run("panics when empty username", func(t *testing.T) {
		c := NewAPIClient(DummyURL)

		assert.Panics(t, func() { _ = c.StartSession("", "123") })
	})
	t.Run("panics when empty password", func(t *testing.T) {
		c := NewAPIClient(DummyURL)

		assert.Panics(t, func() { _ = c.StartSession("test", "") })
	})
	t.Run("adds authToken header to all future HTTP requests", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		c := NewAPIClient(server.URL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}
		wantedAuthToken := "Basic dGVzdDoxMjM="

		// Act
		err := c.StartSession("test", "123")

		// Assert
		assert.NoError(t, err)

		_, _, err = c.Get(DummyRoute)
		require.NoError(t, err)
		assert.Equal(t, wantedAuthToken, spyTransport.Request.Header.Get("Authorization"))

		_, _, err = c.Post(DummyRoute, nil)
		assert.NoError(t, err)
		assert.Equal(t, wantedAuthToken, spyTransport.Request.Header.Get("Authorization"))
	})
	t.Run("adds authorization to websocket upgrade request", func(t *testing.T) {
		// Arrange
		var req *http.Request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req = r
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)
		wantedAuthToken := "Basic dGVzdDoxMjM="

		// Act
		err := client.StartSession("test", "123")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, wantedAuthToken, req.Header.Get("Authorization"), "expected Authorization header to be set")
	})
	t.Run("initializes websocket connection", func(t *testing.T) {
		// Arrange
		var req *http.Request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req = r
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		err := client.StartSession(DummyUsername, DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, req, "upgrade request should have been sent")
		assert.Equal(t, "/ws", req.URL.Path)
	})
	t.Run("handles server url with trailing slash", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		err := client.StartSession(DummyUsername, DummyPassword)

		// Assert
		assert.NoError(t, err)
	})
	t.Run("returns error when connection fails", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "websocket error", http.StatusInternalServerError)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		err := client.StartSession(DummyUsername, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
}

func TestAPIClient_Subscribe(t *testing.T) {
	t.Run("handler is called when message is received", func(t *testing.T) {
		// Arrange
		event := api.MessageTypeSync
		payload := json.RawMessage(`{"value": "test"}`)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
			err = conn.WriteJSON(api.WSMessage{Type: event, Data: payload})
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		received := make(chan json.RawMessage)
		client.Subscribe(event, func(data json.RawMessage) error {
			received <- data
			return nil
		})
		// Start listening after subscribing
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Assert
		select {
		case got := <-received:
			assert.JSONEq(t, string(payload), string(got))
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for message")
		}
	})
	t.Run("multiple handlers for multiple events", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)

			// Message 1
			msg1 := api.WSMessage{Type: "event1", Data: json.RawMessage(`{"data": "1"}`)}
			err = conn.WriteJSON(msg1)
			require.NoError(t, err)
			// Message 2
			msg2 := api.WSMessage{Type: "event2", Data: json.RawMessage(`{"data": "2"}`)}
			err = conn.WriteJSON(msg2)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		event1Received := make(chan struct{})
		event2Received := make(chan struct{})
		client.Subscribe("event1", func(data json.RawMessage) error {
			event1Received <- struct{}{}
			return nil
		})
		client.Subscribe("event2", func(data json.RawMessage) error {
			event2Received <- struct{}{}
			return nil
		})
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Assert
		for i := 0; i < 2; i++ {
			select {
			case <-event1Received:
			case <-event2Received:
			case <-time.After(time.Second):
				assert.Fail(t, "timeout waiting for messages")
			}
		}
	})
	t.Run("handler is not called for different event type", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)

			msg := api.WSMessage{Type: "different_event", Data: json.RawMessage(`{"data": "test"}`)}
			err = conn.WriteJSON(msg)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		handlerCalled := false
		client.Subscribe("test_event", func(data json.RawMessage) error {
			handlerCalled = true
			return nil
		})
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Assert
		time.Sleep(100 * time.Millisecond) // Give some time for potential handler call
		if handlerCalled {
			assert.False(t, handlerCalled, "handler was called for wrong event type")
		}
	})
	t.Run("panics when called after connection was established", func(t *testing.T) {
		// Arrange
		msg := api.WSMessage{Type: "test_event", Data: json.RawMessage(`{"data": "test"}`)}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Act&Assert
		assert.Panics(t, func() { client.Subscribe(msg.Type, func(data json.RawMessage) error { return nil }) })
	})
}

func TestAPIClient_Close(t *testing.T) {
	t.Run("removes authToken header to all future HTTP requests", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		c := NewAPIClient(server.URL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}
		err := c.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Act
		err = c.Close()

		// Assert
		assert.NoError(t, err)
		_, _, err = c.Get(DummyRoute)
		require.NoError(t, err)
		assert.Empty(t, spyTransport.Request.Header.Get("Authorization"), "authToken header should not be set")
		_, _, err = c.Post(DummyRoute, nil)
		require.NoError(t, err)
		assert.Empty(t, spyTransport.Request.Header.Get("Authorization"), "authToken header should not be set")
	})
	t.Run("sends websocket close message to server when connected", func(t *testing.T) {
		// Arrange
		serverConnClosed := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)

			_, _, err = conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					close(serverConnClosed)
				} else {
					assert.Fail(t, "unexpected error: %w", err)
				}
			}
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)

		// Act
		err = client.Close()

		// Assert
		assert.NoError(t, err)
		select {
		case <-serverConnClosed:
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for connection to close")
		}
	})
	t.Run("close is idempotent", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)
		err := client.StartSession(DummyUsername, DummyPassword)
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)

		// Act
		err = client.Close()

		// Assert
		assert.NoError(t, err, "second Close() call should not return error")
	})
}

func TestAPIClient_Get(t *testing.T) {
	t.Run("sends GET request to given url", func(t *testing.T) {
		// Arrange
		serverURL := "http://localhost:5000"
		c := NewAPIClient(serverURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}

		// Act
		_, _, err := c.Get("/test")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.Equal(t, http.MethodGet, spyTransport.Request.Method, "HTTP method should be GET")
		assert.Equal(t, serverURL+"/test", spyTransport.Request.URL.String(), "URL should match")
	})
	t.Run("returns response status and unmarshalls response payload", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		respPayload := TestData{Value: "abc"}
		b, _ := json.Marshal(respPayload)
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.HttpClient = &http.Client{Transport: spyTransport}

		// Act
		status, body, err := c.Get(DummyRoute)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status, "HTTP status should be OK")
		var got TestData
		err = json.Unmarshal(body, &got)
		require.NoError(t, err)
		assert.Equal(t, respPayload, got, "response payload should match")
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _, _ = c.Get("")
		})
	})
}

func TestAPIClient_Post(t *testing.T) {
	t.Run("sends POST request with given payload", func(t *testing.T) {
		// Arrange
		serverURL := "http://localhost:5000"
		c := NewAPIClient(serverURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}
		payload := TestData{Value: "abc"}

		// Act
		_, _, err := c.Post("/test", payload)

		assert.NoError(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.Equal(t, http.MethodPost, spyTransport.Request.Method, "HTTP method should be POST")
		assert.Equal(t, serverURL+"/test", spyTransport.Request.URL.String(), "URL should match")
		assert.Equal(t, "application/json", spyTransport.Request.Header.Get("Content-Type"), "Content-Type header should be set to application/json")
		payloadBytes, _ := json.Marshal(payload)
		gotBytes, _ := io.ReadAll(spyTransport.Request.Body)
		assert.JSONEqf(t, string(payloadBytes), string(gotBytes), "Expected JSON payload to match.")
	})
	t.Run("returns response status and unmarshalls response payload", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		respPayload := TestData{Value: "abc"}
		b, _ := json.Marshal(respPayload)
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.HttpClient = &http.Client{Transport: spyTransport}

		// Act
		status, body, err := c.Post(DummyRoute, nil)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status, "HTTP status should be OK")
		var got TestData
		err = json.Unmarshal(body, &got)
		require.NoError(t, err)
		assert.Equal(t, respPayload, got, "response payload should match")
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _, _ = c.Post("", nil)
		})
	})
}
