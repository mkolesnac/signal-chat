package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"signal-chat/internal/apitypes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPClientSpy is a spy implementation of httpDoer that records requests
type HTTPClientSpy struct {
	requests []*http.Request
	response *http.Response
	err      error
}

func (s *HTTPClientSpy) Do(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	return s.response, s.err
}

// WebsocketClientSpy is a spy implementation of webSocketHandler
type WebsocketClientSpy struct {
	connectCalled bool
	closeCalled   bool
	connectErr    error
	authToken     string
}

func (s *WebsocketClientSpy) Connect(authToken string) error {
	s.connectCalled = true
	s.authToken = authToken
	return s.connectErr
}

func (s *WebsocketClientSpy) Close() {
	s.closeCalled = true
}

func (s *WebsocketClientSpy) SetMessageHandler(messageType apitypes.WSMessageType, handler MessageHandler) {
}

func (s *WebsocketClientSpy) SetConnectionStateHandler(handler ConnectionStateHandler) {
}

func TestNewClient(t *testing.T) {
	t.Run("panics when URL has no protocol", func(t *testing.T) {
		// Act & Assert
		assert.Panics(t, func() {
			NewClient("example.com")
		})
	})

	t.Run("panics when URL is empty", func(t *testing.T) {
		// Act & Assert
		assert.Panics(t, func() {
			NewClient("")
		})
	})
}

func TestClient_AuthHeader(t *testing.T) {
	t.Run("attaches auth header to all future requests after SignUp", func(t *testing.T) {
		// Arrange
		wsSpy := &WebsocketClientSpy{}

		// Simulate sign up
		httpSpy1 := testHTTPClient(t, http.StatusOK, apitypes.SignUpResponse{AuthToken: "test-token"})
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy1,
			wsClient:   wsSpy,
		}
		_, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})
		require.NoError(t, err)

		// Simulate normal server request
		resp := apitypes.GetUserResponse{
			User: apitypes.User{
				ID:       "user123",
				Username: "testuser",
			},
		}
		httpSpy2 := testHTTPClient(t, http.StatusOK, resp)
		client.httpClient = httpSpy2

		// Act
		_, err = client.GetUser("user123")

		// Assert
		require.NoError(t, err)
		require.Len(t, httpSpy2.requests, 1)
		assert.Equal(t, "Bearer test-token", httpSpy2.requests[0].Header.Get("Authorization"))
	})
	t.Run("attaches auth header to all future requests after SignIn", func(t *testing.T) {
		// Arrange
		wsSpy := &WebsocketClientSpy{}

		// Simulate sign up
		httpSpy1 := testHTTPClient(t, http.StatusOK, apitypes.SignInResponse{AuthToken: "test-token"})
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy1,
			wsClient:   wsSpy,
		}
		_, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})
		require.NoError(t, err)

		// Simulate normal server request
		resp := apitypes.GetUserResponse{
			User: apitypes.User{
				ID:       "user123",
				Username: "testuser",
			},
		}
		httpSpy2 := testHTTPClient(t, http.StatusOK, resp)
		client.httpClient = httpSpy2

		// Act
		_, err = client.GetUser("user123")

		// Assert
		require.NoError(t, err)
		require.Len(t, httpSpy2.requests, 1)
		assert.Equal(t, "Bearer test-token", httpSpy2.requests[0].Header.Get("Authorization"))
	})

	t.Run("does not attach auth header when token is empty", func(t *testing.T) {
		// Arrange
		resp := apitypes.GetAllUsersResponse{
			Users: []apitypes.User{},
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.GetAllUsers()

		// Assert
		require.NoError(t, err)
		require.Len(t, httpSpy.requests, 1)
		assert.Empty(t, httpSpy.requests[0].Header.Get("Authorization"))
	})
}

func TestClient_SignUp(t *testing.T) {
	t.Run("successfully signs up and establishes websocket connection", func(t *testing.T) {
		// Arrange
		resp := apitypes.SignUpResponse{
			UserID:    "user123",
			AuthToken: "token123",
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		resp, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "user123", resp.UserID)
		assert.Equal(t, "token123", resp.AuthToken)
		assert.True(t, wsSpy.connectCalled)
		assert.Equal(t, "token123", wsSpy.authToken)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})

		// Assert
		assert.Error(t, err)
		assert.False(t, wsSpy.connectCalled)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "invalid request"}
		httpSpy := testHTTPClient(t, http.StatusBadRequest, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusBadRequest, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
		assert.False(t, wsSpy.connectCalled)
	})

	t.Run("returns error when websocket connection fails", func(t *testing.T) {
		// Arrange
		resp := apitypes.SignUpResponse{
			UserID:    "user123",
			AuthToken: "token123",
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{
			connectErr: errors.New("websocket connection error"),
		}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignUp("testuser", "testpass", apitypes.KeyBundle{RegistrationID: 123})

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, wsSpy.connectErr)
		assert.True(t, wsSpy.connectCalled)
	})
}

func TestClient_SignIn(t *testing.T) {
	t.Run("successfully signs in and establishes websocket connection", func(t *testing.T) {
		// Arrange
		resp := apitypes.SignInResponse{
			UserID:    "user123",
			AuthToken: "token123",
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		resp, err := client.SignIn("testuser", "testpass")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "user123", resp.UserID)
		assert.Equal(t, "token123", resp.AuthToken)
		assert.True(t, wsSpy.connectCalled)
		assert.Equal(t, "token123", wsSpy.authToken)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignIn("testuser", "testpass")

		// Assert
		assert.Error(t, err)
		assert.False(t, wsSpy.connectCalled)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "invalid credentials"}
		httpSpy := testHTTPClient(t, http.StatusUnauthorized, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignIn("testuser", "testpass")

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusUnauthorized, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
		assert.False(t, wsSpy.connectCalled)
	})

	t.Run("returns error when websocket connection fails", func(t *testing.T) {
		// Arrange
		resp := apitypes.SignInResponse{
			UserID:    "user123",
			AuthToken: "token123",
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{
			connectErr: errors.New("websocket connection error"),
		}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
		}

		// Act
		_, err := client.SignIn("testuser", "testpass")

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, wsSpy.connectErr)
		assert.True(t, wsSpy.connectCalled)
	})
}

func TestClient_GetUser(t *testing.T) {
	t.Run("retrieves user by ID successfully", func(t *testing.T) {
		// Arrange
		resp := apitypes.GetUserResponse{
			User: apitypes.User{
				ID:       "user123",
				Username: "testuser",
			},
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		resp, err := client.GetUser("user123")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "user123", resp.User.ID)
		assert.Equal(t, "testuser", resp.User.Username)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.GetUser("user123")

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "user not found"}
		httpSpy := testHTTPClient(t, http.StatusNotFound, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.GetUser("user123")

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusNotFound, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
	})
}

func TestClient_GetAllUsers(t *testing.T) {
	t.Run("retrieves all users successfully", func(t *testing.T) {
		// Arrange
		resp := apitypes.GetAllUsersResponse{
			Users: []apitypes.User{
				{ID: "user123", Username: "user1"},
				{ID: "user234", Username: "user2"},
			},
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		resp, err := client.GetAllUsers()

		// Assert
		require.NoError(t, err)
		assert.Len(t, resp.Users, 2)
		assert.Equal(t, "user123", resp.Users[0].ID)
		assert.Equal(t, "user1", resp.Users[0].Username)
		assert.Equal(t, "user234", resp.Users[1].ID)
		assert.Equal(t, "user2", resp.Users[1].Username)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}

		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.GetAllUsers()

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "unauthorized"}
		httpSpy := testHTTPClient(t, http.StatusUnauthorized, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.GetAllUsers()

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusUnauthorized, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
	})
}

func TestClient_CreateConversation(t *testing.T) {
	t.Run("creates new conversation successfully", func(t *testing.T) {
		// Arrange
		resp := apitypes.CreateConversationResponse{
			ConversationID: "conv123",
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		err := client.CreateConversation("conv123", []apitypes.Participant{{ID: "user1"}})

		// Assert
		require.NoError(t, err)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		err := client.CreateConversation("conv123", []apitypes.Participant{{ID: "user1"}})

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "invalid request"}
		httpSpy := testHTTPClient(t, http.StatusBadRequest, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		err := client.CreateConversation("conv123", []apitypes.Participant{{ID: "user1"}})

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusBadRequest, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
	})
}

func TestClient_SendMessage(t *testing.T) {
	t.Run("sends message to conversation successfully", func(t *testing.T) {
		// Arrange
		resp := apitypes.SendMessageResponse{
			MessageID: "msg123",
			CreatedAt: 1234567890,
		}
		httpSpy := testHTTPClient(t, http.StatusOK, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		resp, err := client.SendMessage("conv123", []byte("Hello, world!"))

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "msg123", resp.MessageID)
		assert.Equal(t, int64(1234567890), resp.CreatedAt)
	})

	t.Run("returns error when HTTP request fails", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			err: errors.New("network error"),
		}
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.SendMessage("conv123", []byte("Hello, world!"))

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns non-OK status", func(t *testing.T) {
		// Arrange
		resp := apitypes.ErrorResponse{Message: "invalid request"}
		httpSpy := testHTTPClient(t, http.StatusBadRequest, resp)
		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		_, err := client.SendMessage("conv123", []byte("Hello, world!"))

		// Assert
		require.Error(t, err)
		var respErr *ServerError
		require.ErrorAs(t, err, &respErr)
		assert.Equal(t, http.StatusBadRequest, respErr.StatusCode)
		assert.Equal(t, resp.Message, respErr.Message)
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("closes websocket connection successfully", func(t *testing.T) {
		// Arrange
		httpSpy := &HTTPClientSpy{
			response: &http.Response{},
		}

		wsSpy := &WebsocketClientSpy{}
		client := &Client{
			ServerURL:  "http://example.com",
			httpClient: httpSpy,
			wsClient:   wsSpy,
			authToken:  "test-token",
		}

		// Act
		client.Close()

		// Assert
		assert.True(t, wsSpy.closeCalled)
	})
}

func testHTTPClient(t *testing.T, statusCode int, response any) *HTTPClientSpy {
	t.Helper()

	body, err := json.Marshal(response)
	require.NoError(t, err)

	return &HTTPClientSpy{
		response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBuffer(body)),
		},
	}
}
