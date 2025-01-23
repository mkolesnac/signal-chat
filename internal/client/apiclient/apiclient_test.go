package apiclient

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type TestData struct {
	Value string
}

const (
	DummyServerURL = "https://dummy.com"
	DummyRoute     = "/dummy"
)

var (
	DummyRoundTripper = &SpyRoundTripper{}
	DummyTarget       = &struct{}{}
)

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
	t.Run("sends GET request to given url", func(t *testing.T) {
		// Arrange
		serverURL := "http://localhost:5000"
		c := NewAPIClient(serverURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}

		// Act
		_, err := c.Get("/test", DummyTarget)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.Equal(t, http.MethodGet, spyTransport.Request.Method, "HTTP method should be GET")
		assert.Equal(t, serverURL+"/test", spyTransport.Request.URL.String(), "URL should match")
	})
	t.Run("attaches authorization header when configured", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}
		c.UseBasicAuth("test", "123")

		// Act
		_, err := c.Get(DummyRoute, DummyTarget)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
	})
	t.Run("returns response status and unmarshalls response payload", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		respPayload := TestData{Value: "abc"}
		b, _ := json.Marshal(respPayload)
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.httpClient = &http.Client{Transport: spyTransport}
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
		c := NewAPIClient(DummyServerURL)
		b := []byte("abc")
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.httpClient = &http.Client{Transport: spyTransport}
		var got TestData

		// Act
		_, err := c.Get(DummyRoute, &got)

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		c.httpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Get("", DummyTarget)
		})
	})
	t.Run("panics when target not pointer", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		c.httpClient = &http.Client{Transport: DummyRoundTripper}
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
		c.httpClient = &http.Client{Transport: spyTransport}
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
	t.Run("attaches authorization header when configured", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		spyTransport := &SpyRoundTripper{}
		c.httpClient = &http.Client{Transport: spyTransport}
		c.UseBasicAuth("test", "123")

		// Act
		_, err := c.Post(DummyRoute, nil, DummyTarget)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Basic dGVzdDoxMjM=", spyTransport.Request.Header.Get("Authorization"))
	})
	t.Run("returns response status and unmarshalls response payload", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		respPayload := TestData{Value: "abc"}
		b, _ := json.Marshal(respPayload)
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.httpClient = &http.Client{Transport: spyTransport}
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
		c := NewAPIClient(DummyServerURL)
		b := []byte("abc")
		resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(b))}
		spyTransport := &SpyRoundTripper{Response: resp}
		c.httpClient = &http.Client{Transport: spyTransport}
		var got TestData

		// Act
		_, err := c.Post(DummyRoute, nil, &got)

		// Assert
		assert.Error(t, err)
	})
	t.Run("panics when empty route", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		c.httpClient = &http.Client{Transport: DummyRoundTripper}

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Post("", nil, DummyTarget)
		})
	})
	t.Run("panics when target not pointer", func(t *testing.T) {
		// Arrange
		c := NewAPIClient(DummyServerURL)
		c.httpClient = &http.Client{Transport: DummyRoundTripper}
		var target TestData

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = c.Post(DummyRoute, nil, target)
		})
	})
}
