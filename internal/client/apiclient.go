package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type APIClient struct {
	ServerURL     string
	Authorization string
	HTTPClient    *http.Client
}

func NewAPIClient(serverURL string) *APIClient {
	return &APIClient{
		ServerURL:  serverURL,
		HTTPClient: &http.Client{},
	}
}

func (a *APIClient) UseBasicAuth(username, password string) {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	a.Authorization = "Basic " + credentials
}

func (a *APIClient) Get(route string) ([]byte, error) {
	url := a.ServerURL + route
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	return a.send(req)
}

func (a *APIClient) Post(route string, payload interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(payload)
	url := a.ServerURL + route
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return a.send(req)
}

func (a *APIClient) send(req *http.Request) ([]byte, error) {
	if a.Authorization != "" {
		req.Header.Set("Authorization", a.Authorization)
	}

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
