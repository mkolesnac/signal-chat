package apiclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

type Client interface {
	Get(route string, target any) (int, error)
	Post(route string, payload any) (int, error)
}

type APIClient struct {
	ServerURL     string
	authorization string
	httpClient    *http.Client
}

func NewAPIClient(serverURL string) *APIClient {
	return &APIClient{
		ServerURL:  serverURL,
		httpClient: &http.Client{},
	}
}

func (a *APIClient) UseBasicAuth(username, password string) {
	requireNonEmpty("username", username)
	requireNonEmpty("password", password)

	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	a.authorization = "Basic " + credentials
}

func (a *APIClient) Get(route string, target any) (int, error) {
	requireNonEmpty("route", route)
	requirePointer("target", target)

	req, err := a.newRequest("GET", route, nil)
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

	req, err := a.newRequest("POST", route, b)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.send(req, target)
}

func (a *APIClient) newRequest(method, route string, payload []byte) (*http.Request, error) {
	url := a.ServerURL + route
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	if a.authorization != "" {
		req.Header.Set("Authorization", a.authorization)
	}

	return req, nil
}

func (a *APIClient) send(req *http.Request, target any) (int, error) {
	resp, err := a.httpClient.Do(req)
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
