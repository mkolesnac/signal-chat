package main

import (
	"encoding/json"
	"errors"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
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
	t.Run("generates key pairs and writes them to db", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		auth := Auth{db: db, apiClient: apiclient.NewFake()}

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.NoError(t, err)
		v, err := db.Read(database.PrivateIdentityKeyPK())
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
		v, err = db.Read(database.PublicIdentityKeyPK())
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
		keys, err := db.Query(database.SignedPreKeyPK(""))
		assert.NoError(t, err)
		assert.Len(t, keys, 1)
		keys, err = db.Query(database.PreKeyPK(""))
		assert.NoError(t, err)
		assert.Len(t, keys, 100)
	})
	t.Run("sends signup request with authorization and credentials to server", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := Auth{db: database.NewFake(), apiClient: ac}

		// Act
		_, err := auth.SignUp("test@user.com", "password123")

		// Assert
		assert.NoError(t, err)

		requests := ac.Requests()
		require.Len(t, requests, 1, "request should have been sent")
		var payload api.SignUpRequest
		err = json.Unmarshal(requests[0].PayloadJSON, &payload)
		require.NoError(t, err)
		assert.NotEmpty(t, requests[0].Headers["Authorization"], "authorization header should be set")
		assert.Equal(t, "test@user.com", payload.UserName)
		assert.Equal(t, "password123", payload.Password)
	})
	t.Run("includes public keys in signup request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := Auth{db: database.NewFake(), apiClient: ac}

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.NoError(t, err)

		requests := ac.Requests()
		require.Len(t, requests, 1, "request should have been sent")
		var payload api.SignUpRequest
		err = json.Unmarshal(requests[0].PayloadJSON, &payload)
		require.NoError(t, err)
		assert.Len(t, payload.IdentityPublicKey, 33, "curve25519 public keys should be 33 bytes long")
		assert.Equal(t, payload.IdentityPublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		assert.Len(t, payload.SignedPreKey.PublicKey, 33, "curve25519 public keys should be 33 bytes long")
		assert.Equal(t, payload.SignedPreKey.PublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		assert.Len(t, payload.SignedPreKey.Signature, 64, "Signed prekey signature should be 64 byte long")
		assert.Len(t, payload.PreKeys, 100, "Request should contain 100 pre keys")
		for _, preKey := range payload.PreKeys {
			assert.Len(t, preKey.PublicKey, 33, "curve25519 public keys should be 33 bytes long")
			assert.Equal(t, preKey.PublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		}
	})
	t.Run("returns registered user on successful response from server", func(t *testing.T) {
		// Arrange
		auth := Auth{db: database.NewFake(), apiClient: apiclient.NewFake()}

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
		auth := Auth{db: db, apiClient: apiclient.NewFake()}

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when database fails to write", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.WriteErr = errors.New("write error")
		auth := Auth{db: db, apiClient: apiclient.NewFake()}

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostErr = errors.New("post error")
		auth := Auth{db: database.NewFake(), apiClient: ac}

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostStatus = http.StatusInternalServerError
		auth := Auth{db: database.NewFake(), apiClient: ac}

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
		auth := Auth{db: db, apiClient: apiclient.NewFake()}
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		_, err = auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.NotPanics(t, func() { _, _ = db.Read(database.PrivateIdentityKeyPK()) }, "Read should not panic if database connection was opened")
	})
	t.Run("sends sign in request with user credentials to server", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := Auth{db: database.NewFake(), apiClient: ac}
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
		auth := Auth{db: database.NewFake(), apiClient: apiclient.NewFake()}
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
		auth := Auth{db: db, apiClient: apiclient.NewFake()}

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when database fails to write", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.WriteErr = errors.New("write error")
		auth := Auth{db: db, apiClient: apiclient.NewFake()}

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostErr = errors.New("post error")
		auth := Auth{db: database.NewFake(), apiClient: ac}

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when user doesn't exist", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostStatus = http.StatusInternalServerError
		auth := Auth{db: database.NewFake(), apiClient: ac}

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
		auth := Auth{db: db, apiClient: apiclient.NewFake()}
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.NoError(t, err)
		assert.Panics(t, func() { _, _ = db.Read(database.PrivateIdentityKeyPK()) }, "read should panic because database connection was closed")
	})
	t.Run("resets authorization header on api client", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := Auth{db: database.NewFake(), apiClient: ac}
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.NoError(t, err)
		_, err = ac.Post("/test", struct{}{}, nil)
		assert.NoError(t, err)
		requests := ac.Requests()
		require.Len(t, requests, 2, "sign out request should have been sent after signup request")
		assert.NotContains(t, requests[1].Headers, "Authorization", "authorization header should not be attached to requests after sign out")
	})
	t.Run("returns error when database fails to close", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.CloseErr = errors.New("close error")
		auth := Auth{db: db, apiClient: apiclient.NewFake()}
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.Error(t, err)
	})
}
