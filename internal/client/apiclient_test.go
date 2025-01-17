package client

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

const DummyServerURL = "http://localhost:5000"

func TestAPIClient_UseBasicAuth(t *testing.T) {
	t.Run("panics when empty username", func(t *testing.T) {
		c := NewAPIClient(DummyServerURL)

		assert.Panics(t, func() { c.UseBasicAuth("", "123") })
	})
	t.Run("panics when empty password", func(t *testing.T) {
		c := NewAPIClient(DummyServerURL)

		assert.Panics(t, func() { c.UseBasicAuth("test", "") })
	})
}

func TestAPIClient_Get(t *testing.T) {
	t.Run("sends request with authorization header", func(t *testing.T) {
		c := NewAPIClient(DummyServerURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}

		c.UseBasicAuth("test", "123")
		_, err := c.Get("/test")

		assert.Nil(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.Equal(t, http.MethodGet, spyTransport.Request.Method, "HTTP method should be GET")
		assert.Equal(t, DummyServerURL+"/test", spyTransport.Request.URL.String(), "URL should match")
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
	})
}

func TestAPIClient_Post(t *testing.T) {
	t.Run("sends request with JSON body", func(t *testing.T) {
		c := NewAPIClient(DummyServerURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}

		payload := struct {
			value string
		}{
			value: "abc",
		}
		_, err := c.Post("/test", payload)

		assert.Nil(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.Equal(t, http.MethodPost, spyTransport.Request.Method, "HTTP method should be POST")
		assert.Equal(t, DummyServerURL+"/test", spyTransport.Request.URL.String(), "URL should match")
		assert.Equal(t, "application/json", spyTransport.Request.Header.Get("Content-Type"), "Content-Type header should be set to application/json")

		wantPayloadBytes, _ := json.Marshal(payload)
		gotPayloadBytes, _ := io.ReadAll(spyTransport.Request.Body)
		assert.JSONEqf(t, string(wantPayloadBytes), string(gotPayloadBytes), "Expected JSON payload to match.\nWant: %s\nGot: %s", wantPayloadBytes, gotPayloadBytes)
	})
	t.Run("sends POST request with authorization", func(t *testing.T) {
		c := NewAPIClient(DummyServerURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}

		c.UseBasicAuth("test", "123")
		_, err := c.Post("/test", struct{}{})

		assert.Nil(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
	})
}
