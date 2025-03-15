package main

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"signal-chat/internal/api"
	"testing"
)

const (
	DummyEmail    = "dummy@gmail.com"
	DummyPassword = "dummy123"
)

func TestAuth_SignUp(t *testing.T) {
	t.Run("returns error if email is invalid", func(t *testing.T) {
		auth := Auth{}
		_, err := auth.SignUp("", DummyPassword)

		assert.ErrorIs(t, err, ErrAuthInvalidEmail)
	})
	t.Run("returns error if password is shorter than 8 characters", func(t *testing.T) {
		auth := Auth{}
		_, err := auth.SignUp(DummyEmail, "")

		assert.ErrorIs(t, err, ErrAuthPwdTooShort)
	})
	t.Run("sends signup request with credentials and key bundle to server", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		bundle := api.KeyBundle{
			RegistrationID: 0,
			IdentityKey:    randomBytes(32),
			SignedPreKey: api.SignedPreKey{
				ID:        0,
				PublicKey: randomBytes(32),
				Signature: randomBytes(64),
			},
			PreKeys: []api.PreKey{{
				ID:        0,
				PublicKey: randomBytes(32),
			}},
		}
		encrypt := &encryption.ManagerStub{
			InitializeKeyStoreResult: bundle,
		}
		auth := NewAuth(database.NewFake(), ac, encrypt)

		// Act
		_, err := auth.SignUp("test@user.com", "password123")

		// Assert
		assert.NoError(t, err)

		requests := ac.Requests()
		require.Len(t, requests, 1, "request should have been sent")
		var payload api.SignUpRequest
		err = json.Unmarshal(requests[0].PayloadJSON, &payload)
		require.NoError(t, err)
		assert.Equal(t, "test@user.com", payload.UserName)
		assert.Equal(t, "password123", payload.Password)
		assert.Equal(t, bundle, payload.KeyBundle)
	})
	t.Run("returns registered user on successful response from server", func(t *testing.T) {
		// Arrange
		auth := NewAuth(database.NewFake(), apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		user, err := auth.SignUp("test@user.com", DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, user.Username, "test@user.com")
		assert.NotEmpty(t, user.ID)
	})
	t.Run("returns error when database fails to open", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.OpenErr = errors.New("open error")
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostErrors[api.EndpointSignUp] = errors.New("post error")
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when encryption key bundle fails to be generated", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.OpenErr = errors.New("open error")
		en := &encryption.ManagerStub{
			InitializeKeyStoreError: errors.New("test error"),
		}
		auth := NewAuth(db, apiclient.NewFake(), en)

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointSignUp] = apiclient.StubResponse{
			StatusCode: http.StatusInternalServerError,
		}
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns invalid response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointSignUp] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       []byte("invalid response"),
		}
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
}

func TestAuth_SignIn(t *testing.T) {
	t.Run("returns error if email is invalid", func(t *testing.T) {
		// Arrange
		auth := Auth{}

		// Act
		_, err := auth.SignIn("", DummyPassword)

		// Assert
		assert.ErrorIs(t, err, ErrAuthInvalidEmail)
	})
	t.Run("returns error if password is shorter than 8 characters", func(t *testing.T) {
		// Arrange
		auth := Auth{}

		// Act
		_, err := auth.SignIn(DummyEmail, "")

		assert.ErrorIs(t, err, ErrAuthPwdTooShort)
	})
	t.Run("can read user data after opening database", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		_, err = auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.NotPanics(t, func() { _, _ = db.Read("test") }, "Read should not panic if database connection was opened")
	})
	t.Run("sends sign in request with user credentials to server", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())
		email := "test@user.com"
		pwd := "password123"
		_, err := auth.SignUp(email, pwd)
		require.NoError(t, err)

		// Act
		_, err = auth.SignIn(email, pwd)

		// Assert
		assert.NoError(t, err)

		requests := ac.Requests()
		require.Len(t, requests, 2, "sign request should have been sent after signup request")
		assert.Equal(t, api.EndpointSignIn, requests[1].Route, "request should have been sent to /signin route")
		assert.NotEmpty(t, requests[1].Headers["Authorization"], "authorization header should be set")

		var payload api.SignInRequest
		err = json.Unmarshal(requests[0].PayloadJSON, &payload)
		require.NoError(t, err)
		assert.Equal(t, email, payload.Username)
		assert.Equal(t, pwd, payload.Password)
	})
	t.Run("returns registered user on successful response from server", func(t *testing.T) {
		// Arrange
		auth := NewAuth(database.NewFake(), apiclient.NewFake(), encryption.NewManagerFake())
		username := "test@user.com"
		registered, err := auth.SignUp(username, DummyPassword)
		require.NoError(t, err)

		// Act
		signedIn, err := auth.SignIn(username, DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, registered.ID, signedIn.ID, "user ID should match the ID of the user returned from SignUp")
		assert.Equal(t, username, signedIn.Username, "username should match the username of the user returned from SignUp")
	})
	t.Run("returns error when database fails to open", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.OpenErr = errors.New("open error")
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when database fails to write", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.WriteErr = errors.New("write error")
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostErrors[api.EndpointSignIn] = errors.New("test error")
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointSignIn] = apiclient.StubResponse{
			StatusCode: http.StatusInternalServerError,
		}
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns invalid response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostResponses[api.EndpointSignIn] = apiclient.StubResponse{
			StatusCode: http.StatusOK,
			Body:       []byte("invalid response"),
		}
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
}

func TestAuth_SignOut(t *testing.T) {
	t.Run("panics if not signed in", func(t *testing.T) {
		auth := Auth{}
		assert.Panics(t, func() { _ = auth.SignOut() }, "should panic when no user is signed in")
	})
	t.Run("closes database connection", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.NoError(t, err)
		assert.Panics(t, func() { _, _ = db.Read("test") }, "read should panic because database connection was closed")
	})
	t.Run("resets authorization header on api client", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := NewAuth(database.NewFake(), ac, encryption.NewManagerFake())
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.NoError(t, err)
		_, _, err = ac.Post("/test", struct{}{})
		assert.NoError(t, err)
		requests := ac.Requests()
		require.Len(t, requests, 2, "sign out request should have been sent after signup request")
		assert.NotContains(t, requests[1].Headers, "Authorization", "authorization header should not be attached to requests after sign out")
	})
	t.Run("returns error when database fails to close", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.CloseErr = errors.New("close error")
		auth := NewAuth(db, apiclient.NewFake(), encryption.NewManagerFake())
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.Error(t, err)
	})
}
