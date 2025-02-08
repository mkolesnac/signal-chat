package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"reflect"
	"signal-chat/internal/api"
	"strings"
	"time"
)

type HTTPClient interface {
	// Get performs a GET request and unmarshals the response into target.
	// The target must be a pointer to a value that can be unmarshalled from JSON.
	Get(route string, target any) (int, error)

	// Post performs a POST request with the given payload and unmarshals the response into target.
	// The target must be a pointer to a value that can be unmarshalled from JSON.
	Post(route string, payload any, target any) (int, error)
}

type EventSubscriber interface {
	Subscribe(eventType string, handler func([]byte))
}

type Authenticator interface {
	SetAuthorization(username, password string)
	ClearAuthorization()
}

type APIClient struct {
	ServerURL  string
	authToken  string
	HttpClient *http.Client
	WSConn     *websocket.Conn
	handlers   map[string]func([]byte)
}

func NewAPIClient(serverURL string) *APIClient {
	requireNonEmpty("serverURL", serverURL)
	if !strings.Contains(serverURL, "http://") && !strings.Contains(serverURL, "https://") {
		panic("serverURL must start with either http:// or https://")
	}

	return &APIClient{
		ServerURL:  strings.TrimSuffix(serverURL, "/"),
		HttpClient: &http.Client{},
		handlers:   make(map[string]func([]byte)),
	}
}

func (a *APIClient) SetAuthorization(username, password string) {
	requireNonEmpty("username", username)
	requireNonEmpty("password", password)

	a.authToken = basicAuthorization(username, password)
}

func (a *APIClient) ClearAuthorization() {
	a.authToken = ""
}

func (a *APIClient) Get(route string, target any) (int, error) {
	requireNonEmpty("route", route)
	requirePointer("target", target)

	req, err := a.newHTTPRequest("GET", route, nil)
	if err != nil {
		return 0, err
	}

	return a.send(req, target)
}

func (a *APIClient) Post(route string, payload any, target any) (int, error) {
	requireNonEmpty("route", route)
	requirePointer("target", target)

	b, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := a.newHTTPRequest("POST", route, b)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.send(req, target)
}

func (a *APIClient) StartListening() error {
	header := http.Header{}
	if a.authToken != "" {
		header.Add("Authorization", a.authToken)
	}

	serverURL := strings.TrimPrefix(a.ServerURL, "http://")
	serverURL = strings.TrimPrefix(serverURL, "https://")
	if strings.Contains(a.ServerURL, "https://") {
		serverURL = "wss://" + serverURL
	} else {
		serverURL = "ws://" + serverURL
	}

	conn, _, err := websocket.DefaultDialer.Dial(serverURL+"/ws", header)
	if err != nil {
		return err
	}

	a.WSConn = conn
	go a.listenWebSocket()

	return nil
}

func (a *APIClient) Subscribe(eventType string, handler func([]byte)) {
	if a.WSConn != nil {
		panic("cannot subscribe to new events after the websocket connection has been opened")
	}

	a.handlers[eventType] = handler
}

func (a *APIClient) Close() error {
	if a.WSConn != nil {
		// Send close message
		message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		err := a.WSConn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		if err != nil {
			return err
		}

		// Close the connection
		if err := a.WSConn.Close(); err != nil {
			return err
		}

		a.WSConn = nil
	}

	return nil
}

func (a *APIClient) reconnect() error {
	maxRetries := 5
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		err := a.StartListening()
		if err == nil {
			return nil
		}

		time.Sleep(backoff)
		backoff *= 2 // exponential backoff
	}

	return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
}

func (a *APIClient) listenWebSocket() {
	defer func() {
		if a.WSConn != nil {
			_ = a.WSConn.Close()
		}
	}()

	for {
		var event api.WSMessage
		err := a.WSConn.ReadJSON(&event)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return // Exit on non-reconnectable errors
			}
			if err := a.reconnect(); err != nil {
				log.Printf("Failed to reconnect: %v", err)
				return
			}
			continue
		}

		handler, exists := a.handlers[event.Type]

		if exists {
			handler(event.Payload)
		}
	}
}

func (a *APIClient) newHTTPRequest(method, route string, payload []byte) (*http.Request, error) {
	url := a.ServerURL + route
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	if a.authToken != "" {
		req.Header.Set("Authorization", a.authToken)
	}

	return req, nil
}

func (a *APIClient) send(req *http.Request, target any) (int, error) {
	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return resp.StatusCode, nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.StatusCode, nil
}

func requireNonEmpty(name, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}

func requirePointer(name string, value any) {
	if value == nil {
		panic(fmt.Sprintf("%s cannot be nil", name))
	}

	// Check if target is a pointer
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("target must be a pointer, got %T", value))
	}
}
