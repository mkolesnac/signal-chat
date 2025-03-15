package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"signal-chat/internal/api"
	"strings"
	"sync"
	"time"
)

type Client interface {
	Get(route string) (int, []byte, error)
	Post(route string, payload any) (int, []byte, error)
}

type APIClient struct {
	ServerURL  string
	authToken  string
	HttpClient *http.Client
	WSConn     *websocket.Conn
	wsLock     sync.Mutex
	handlers   map[api.WSMessageType]api.WSHandler
	onWSError  api.WSErrorHandler
}

func NewAPIClient(serverURL string) *APIClient {
	requireNonEmpty("serverURL", serverURL)
	if !strings.Contains(serverURL, "http://") && !strings.Contains(serverURL, "https://") {
		panic("serverURL must start with either http:// or https://")
	}

	return &APIClient{
		ServerURL:  strings.TrimSuffix(serverURL, "/"),
		HttpClient: &http.Client{},
		handlers:   make(map[api.WSMessageType]api.WSHandler),
	}
}

func (a *APIClient) StartSession(username, password string) error {
	requireNonEmpty("username", username)
	requireNonEmpty("password", password)

	a.authToken = basicAuthorization(username, password)

	if err := a.connectWS(); err != nil {
		return err
	}

	return nil
}

func (a *APIClient) Close() error {
	a.authToken = ""

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

func (a *APIClient) SetErrorHandler(handler api.WSErrorHandler) {
	a.onWSError = handler
}

func (a *APIClient) Get(route string) (int, []byte, error) {
	requireNonEmpty("route", route)

	req, err := a.newHTTPRequest("GET", route, nil)
	if err != nil {
		return 0, nil, err
	}

	return a.sendHTTP(req)
}

func (a *APIClient) Post(route string, payload any) (int, []byte, error) {
	requireNonEmpty("route", route)

	b, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := a.newHTTPRequest("POST", route, b)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.sendHTTP(req)
}

func (a *APIClient) Subscribe(eventType api.WSMessageType, handler api.WSHandler) {
	if a.WSConn != nil {
		panic("cannot subscribe to new events after the websocket connection has been opened")
	}

	a.handlers[eventType] = handler
}

func (a *APIClient) connectWS() error {
	header := http.Header{"Authorization": []string{a.authToken}}
	wsURL := strings.Replace(strings.Replace(a.ServerURL, "https://", "wss://", 1), "http://", "ws://", 1) + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return err
	}
	a.WSConn = conn
	go a.listenWebSocket()
	return nil
}

func (a *APIClient) reconnectWS() error {
	maxRetries := 5
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		if err := a.connectWS(); err == nil {
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
		var msg api.WSMessage
		err := a.WSConn.ReadJSON(&msg)
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return // Exit on non-reconnectable errors
			}
			if err := a.reconnectWS(); err != nil {
				a.handleWSError(fmt.Errorf("websocket reconnection failed: %w", err))
				return
			}
			continue
		}

		if handler, exists := a.handlers[msg.Type]; exists {
			err = handler(msg.Data)
			if err != nil {
				a.handleWSError(err)
			} else {
				// Only send ACK after successful processing
				if err := a.sendAcknowledgement(msg); err != nil {
					a.handleWSError(fmt.Errorf("failed to send acknowledgement: %w", err))
				}
			}
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

func (a *APIClient) sendHTTP(req *http.Request) (int, []byte, error) {
	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, body, nil
}

func requireNonEmpty(name, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}

func (a *APIClient) handleWSError(err error) {
	if a.onWSError != nil {
		a.onWSError(err)
	} else {
		panic(fmt.Errorf("unhlandled websocket error: %w", err))
	}
}

func (a *APIClient) sendAcknowledgement(original api.WSMessage) error {
	ack := api.WSAcknowledgementPayload{
		ID:        original.ID,
		Type:      original.Type,
		Timestamp: time.Now().UnixMilli(),
	}
	data, err := json.Marshal(ack)
	if err != nil {
		return fmt.Errorf("failed to marshal acknowledgement: %w", err)
	}

	msg := api.WSMessage{
		ID:   uuid.NewString(), // Generate a new ID for the ACK message
		Type: api.MessageTypeAck,
		Data: data,
	}

	a.wsLock.Lock() // Add a mutex to protect concurrent writes
	defer a.wsLock.Unlock()

	if a.WSConn == nil {
		return fmt.Errorf("websocket connection is not established")
	}

	return a.WSConn.WriteJSON(msg)
}
