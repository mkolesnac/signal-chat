package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"signal-chat/internal/apitypes"
	"strings"
)

type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server returned unsuccessful response: %d - %s", e.StatusCode, e.Message)
}

// httpDoer defines the interface for HTTP operations
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// webSocketHandler defines the interface for WebSocket operations
type webSocketHandler interface {
	Connect(authToken string) error
	Close()
	SetMessageHandler(messageType apitypes.WSMessageType, handler MessageHandler)
	SetConnectionStateHandler(handler ConnectionStateHandler)
}

type Client struct {
	ServerURL  string
	authToken  string
	httpClient httpDoer
	wsClient   webSocketHandler
}

func NewClient(serverURL string) *Client {
	panicIfEmpty("serverURL", serverURL)
	if !strings.Contains(serverURL, "http://") && !strings.Contains(serverURL, "https://") {
		panic("serverURL must start with either http:// or https://")
	}

	trimmed := strings.TrimSuffix(serverURL, "/")
	return &Client{
		ServerURL:  trimmed,
		httpClient: &http.Client{},
		wsClient:   NewWebSocketClient(trimmed),
	}
}

func (c *Client) SetWSMessageHandler(eventType apitypes.WSMessageType, handler MessageHandler) {
	c.wsClient.SetMessageHandler(eventType, handler)
}

func (c *Client) SetConnectionStateHandler(handler ConnectionStateHandler) {
	c.wsClient.SetConnectionStateHandler(handler)
}

func (c *Client) SignUp(username, password string, keyBundle apitypes.KeyBundle) (apitypes.SignUpResponse, error) {
	panicIfEmpty("username", username)
	panicIfEmpty("password", password)

	req := apitypes.SignUpRequest{
		Username:  username,
		Password:  password,
		KeyBundle: keyBundle,
	}

	status, body, err := c.post(apitypes.EndpointSignUp, req)
	if err != nil {
		return apitypes.SignUpResponse{}, err
	}
	if status != http.StatusOK {
		return apitypes.SignUpResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.SignUpResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.SignUpResponse{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}
	c.authToken = resp.AuthToken

	if err := c.wsClient.Connect(c.authToken); err != nil {
		return apitypes.SignUpResponse{}, fmt.Errorf("failed to establish websocket connection: %w", err)
	}

	return resp, nil
}

func (c *Client) SignIn(username, password string) (apitypes.SignInResponse, error) {
	panicIfEmpty("username", username)
	panicIfEmpty("password", password)

	req := apitypes.SignInRequest{
		Username: username,
		Password: password,
	}

	status, body, err := c.post(apitypes.EndpointSignIn, req)
	if err != nil {
		return apitypes.SignInResponse{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return apitypes.SignInResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.SignInResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.SignInResponse{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}
	c.authToken = resp.AuthToken

	if err := c.wsClient.Connect(c.authToken); err != nil {
		return apitypes.SignInResponse{}, fmt.Errorf("failed to establish websocket connection: %w", err)
	}

	return resp, nil
}

func (c *Client) Close() {
	c.authToken = ""
	c.wsClient.Close()
}

func (c *Client) GetUser(id string) (apitypes.GetUserResponse, error) {
	panicIfEmpty("id", id)

	path := strings.Replace(apitypes.EndpointUser, ":id", id, 1)
	status, body, err := c.get(path)
	if err != nil {
		return apitypes.GetUserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}
	if status != http.StatusOK {
		return apitypes.GetUserResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.GetUserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.GetUserResponse{}, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return resp, nil
}

func (c *Client) GetAllUsers() (apitypes.GetAllUsersResponse, error) {
	status, body, err := c.get(apitypes.EndpointUser)
	if err != nil {
		return apitypes.GetAllUsersResponse{}, fmt.Errorf("failed to get user: %w", err)
	}
	if status != http.StatusOK {
		return apitypes.GetAllUsersResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.GetAllUsersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.GetAllUsersResponse{}, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return resp, nil
}

func (c *Client) GetPreKeyBundle(id string) (apitypes.GetPreKeyBundleResponse, error) {
	panicIfEmpty("id", id)

	path := strings.Replace(apitypes.EndpointPreKeyBundle, ":id", id, 1)
	status, body, err := c.get(path)
	if err != nil {
		return apitypes.GetPreKeyBundleResponse{}, fmt.Errorf("failed to get pre key bundle for user %s: %w", id, err)
	}
	if status != http.StatusOK {
		return apitypes.GetPreKeyBundleResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.GetPreKeyBundleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.GetPreKeyBundleResponse{}, fmt.Errorf("failed to unmarshal prekey bundle response: %w", err)
	}

	return resp, nil
}

func (c *Client) CreateConversation(id string, otherParticipants []apitypes.Participant) error {
	panicIfEmpty("id", id)
	if len(otherParticipants) == 0 {
		panic("cannot create conversation without any participants")
	}

	req := apitypes.CreateConversationRequest{
		ConversationID:    id,
		OtherParticipants: otherParticipants,
	}

	status, body, err := c.post(apitypes.EndpointConversations, req)
	if err != nil {
		return fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return parseResponseError(status, body)
	}

	var resp apitypes.CreateConversationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	return nil
}

func (c *Client) SendMessage(conversationID string, content []byte) (apitypes.SendMessageResponse, error) {
	panicIfEmpty("conversationID", conversationID)
	if content == nil || len(content) == 0 {
		panic("content must not be nil or empty")
	}

	req := apitypes.SendMessageRequest{
		ConversationID: conversationID,
		Content:        content,
	}

	status, body, err := c.post(apitypes.EndpointMessages, req)
	if err != nil {
		return apitypes.SendMessageResponse{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return apitypes.SendMessageResponse{}, parseResponseError(status, body)
	}

	var resp apitypes.SendMessageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apitypes.SendMessageResponse{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	return resp, nil
}

func (c *Client) get(route string) (int, []byte, error) {
	panicIfEmpty("route", route)

	req, err := c.newHTTPRequest("GET", route, nil)
	if err != nil {
		return 0, nil, err
	}

	return c.sendHTTP(req)
}

func (c *Client) post(route string, payload any) (int, []byte, error) {
	panicIfEmpty("route", route)

	b, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := c.newHTTPRequest("POST", route, b)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.sendHTTP(req)
}

func (c *Client) newHTTPRequest(method, route string, payload []byte) (*http.Request, error) {
	url := c.ServerURL + route
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	return req, nil
}

func (c *Client) sendHTTP(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, body, nil
}

func panicIfEmpty(name, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", name))
	}
}

func parseResponseError(status int, body []byte) error {
	var errResp apitypes.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("got error unmarshalling error response from server: %w", err)
	}

	return &ServerError{
		StatusCode: status,
		Message:    errResp.Message,
	}
}
