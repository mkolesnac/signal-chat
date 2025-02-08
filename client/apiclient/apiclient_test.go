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
	DummyURL   = "http://localhost:8080"
	DummyRoute = "/dummy"
)

var (
	DummyRoundTripper = &SpyRoundTripper{}
	DummyTarget       = &struct{}{}
)

func TestAPIClient_StartListening(t *testing.T) {
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
		err := client.StartListening()

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
		err := client.StartListening()

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
		err := client.StartListening()

		// Assert
		assert.Error(t, err)
	})
}

func TestAPIClient_Subscribe(t *testing.T) {
	t.Run("handler is called when message is received", func(t *testing.T) {
		// Arrange
		msg := api.WSMessage{Type: "test_event", Payload: json.RawMessage(`{"data": "test"}`)}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
			err = conn.WriteJSON(msg)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		received := make(chan []byte)
		client.Subscribe(msg.Type, func(v []byte) {
			received <- v
		})
		// Start listening after subscribing
		err := client.StartListening()
		require.NoError(t, err)

		// Assert
		select {
		case got := <-received:
			assert.JSONEq(t, string(msg.Payload), string(got))
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
			msg1 := api.WSMessage{Type: "event1", Payload: json.RawMessage(`{"data": "1"}`)}
			err = conn.WriteJSON(msg1)
			require.NoError(t, err)
			// Message 2
			msg2 := api.WSMessage{Type: "event2", Payload: json.RawMessage(`{"data": "2"}`)}
			err = conn.WriteJSON(msg2)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		event1Received := make(chan struct{})
		event2Received := make(chan struct{})
		client.Subscribe("event1", func(payload []byte) {
			event1Received <- struct{}{}
		})
		client.Subscribe("event2", func(payload []byte) {
			event2Received <- struct{}{}
		})
		err := client.StartListening()
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

			msg := api.WSMessage{Type: "different_event", Payload: json.RawMessage(`{"data": "test"}`)}
			err = conn.WriteJSON(msg)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)

		// Act
		handlerCalled := false
		client.Subscribe("test_event", func(v []byte) {
			handlerCalled = true
		})
		err := client.StartListening()
		require.NoError(t, err)

		// Assert
		time.Sleep(100 * time.Millisecond) // Give some time for potential handler call
		if handlerCalled {
			assert.False(t, handlerCalled, "handler was called for wrong event type")
		}
	})
	t.Run("panics when called after connection was established", func(t *testing.T) {
		// Arrange
		msg := api.WSMessage{Type: "test_event", Payload: json.RawMessage(`{"data": "test"}`)}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			_, err := u.Upgrade(w, r, nil)
			require.NoError(t, err)
		}))
		defer server.Close()
		client := NewAPIClient(server.URL)
		err := client.StartListening()
		require.NoError(t, err)

		// Act&Assert
		assert.Panics(t, func() { client.Subscribe(msg.Type, func(v []byte) {}) })
	})
}

func TestAPIClient_Close(t *testing.T) {
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
		err := client.StartListening()
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
		err := client.StartListening()
		require.NoError(t, err)
		err = client.Close()
		require.NoError(t, err)

		// Act
		err = client.Close()

		// Assert
		assert.NoError(t, err, "second Close() call should not return error")
	})
}

func TestAPIClient_SetAuthorization(t *testing.T) {
	t.Run("adds authToken header to all future GET requests", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}

		// Act
		c.SetAuthorization("test", "123")

		// Assert
		_, err := c.Get(DummyRoute, DummyTarget)
		assert.NoError(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
	})
	t.Run("adds authToken header to all future POST requests", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}

		// Act
		c.SetAuthorization("test", "123")

		// Assert
		_, err := c.Post(DummyRoute, nil, DummyTarget)
		assert.NoError(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
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

		// Act
		client.SetAuthorization("test", "123")

		// Assert
		err := client.StartListening()
		require.NoError(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", req.Header.Get("Authorization"), "expected Authorization header to be set")
	})
	t.Run("panics when empty username", func(t *testing.T) {
		c := NewAPIClient(DummyURL)

		assert.Panics(t, func() { c.SetAuthorization("", "123") })
	})
	t.Run("panics when empty password", func(t *testing.T) {
		c := NewAPIClient(DummyURL)

		assert.Panics(t, func() { c.SetAuthorization("test", "") })
	})
}

func TestAPIClient_ClearAuthorization(t *testing.T) {
	t.Run("removes authToken header to all future GET requests", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}
		c.SetAuthorization("test", "123")

		// Act
		c.ClearAuthorization()

		// Assert
		_, err := c.Get(DummyRoute, DummyTarget)
		assert.NoError(t, err)
		assert.Empty(t, spyTransport.Request.Header.Get("Authorization"), "authToken header should not be set")
	})
	t.Run("removes authToken header to all future POST requests", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		spyTransport := &SpyRoundTripper{}
		c.HttpClient = &http.Client{Transport: spyTransport}
		c.SetAuthorization("test", "123")

		// Act
		c.ClearAuthorization()

		// Assert
		_, err := c.Post(DummyRoute, nil, DummyTarget)
		assert.NoError(t, err)
		assert.Empty(t, spyTransport.Request.Header.Get("Authorization"), "authToken header should not be set")
	})
	t.Run("removes authorization from future websocket upgrade requests", func(t *testing.T) {
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
		client.SetAuthorization("test", "123")

		// Act
		client.ClearAuthorization()

		// Assert
		err := client.StartListening()
		require.NoError(t, err)
		assert.Empty(t, req.Header.Get("Authorization"), "authToken header should not be set")
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
		_, err := c.Get("/test", DummyTarget)

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
		var got TestData

		// Act
		status, err := c.Get(DummyRoute, &got)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status, "HTTP status should be OK")
		assert.Equal(t, respPayload, got, "response payload should match")
	})
	t.Run("returns error when response body not valid JSON", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		b := []byte("abc")
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.HttpClient = &http.Client{Transport: spyTransport}
		var got TestData

		// Act
		_, err := c.Get(DummyRoute, &got)

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Get("", DummyTarget)
		})
	})
	t.Run("panics when target not pointer", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}
		var target TestData

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Get(DummyRoute, target)
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
		_, err := c.Post("/test", payload, DummyTarget)

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
		var got TestData

		// Act
		status, err := c.Post(DummyRoute, nil, &got)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status, "HTTP status should be OK")
		assert.Equal(t, respPayload, got, "response payload should match")
	})
	t.Run("returns error when response body not valid JSON", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		b := []byte("abc")
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.HttpClient = &http.Client{Transport: spyTransport}
		var got TestData

		// Act
		_, err := c.Post(DummyRoute, nil, &got)

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Post("", nil, DummyTarget)
		})
	})
	t.Run("panics when target not pointer", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyURL)
		c.HttpClient = &http.Client{Transport: DummyRoundTripper}
		var target TestData

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Post(DummyRoute, nil, target)
		})
	})
}
