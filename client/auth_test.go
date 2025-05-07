package main

import (
	"errors"
	"net/http"
	"signal-chat/client/api"
	"signal-chat/client/database"
	"signal-chat/client/encryption"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("returns registered user on successful response from server", func(t *testing.T) {
		// Arrange
		auth := NewAuth(database.NewFake(), api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		user, err := auth.SignUp("test@user.com", DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test@user.com", user.Username)
		assert.NotEmpty(t, user.ID)
	})

	t.Run("returns error when database fails to open", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.OpenErr = errors.New("open error")
		auth := NewAuth(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		client := api.NewStubClient()
		client.SignUpError = errors.New("post error")
		auth := NewAuth(database.NewFake(), client, encryption.NewFakeManager())

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when encryption key bundle fails to be generated", func(t *testing.T) {
		// Arrange
		en := &encryption.StubManager{
			InitializeKeyStoreError: errors.New("test error"),
		}
		auth := NewAuth(database.NewFake(), api.NewFakeClient(), en)

		// Act
		_, err := auth.SignUp(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		client := api.NewStubClient()
		client.SignUpError = &api.ServerError{
			StatusCode: http.StatusInternalServerError,
		}
		auth := NewAuth(database.NewFake(), client, encryption.NewFakeManager())

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
		auth := NewAuth(db, api.NewFakeClient(), encryption.NewFakeManager())
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		_, err = auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.NoError(t, err)
		assert.NotPanics(t, func() { _, _ = db.Read("test") }, "Read should not panic if database connection was opened")
	})

	t.Run("returns registered user on successful response from server", func(t *testing.T) {
		// Arrange
		auth := NewAuth(database.NewFake(), api.NewFakeClient(), encryption.NewFakeManager())
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
		auth := NewAuth(db, api.NewFakeClient(), encryption.NewFakeManager())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		client := api.NewStubClient()
		client.SignInError = errors.New("test error")
		auth := NewAuth(database.NewFake(), client, encryption.NewFakeManager())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})

	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		client := api.NewStubClient()
		client.SignInError = &api.ServerError{
			StatusCode: http.StatusInternalServerError,
		}
		auth := NewAuth(database.NewFake(), client, encryption.NewFakeManager())

		// Act
		_, err := auth.SignIn(DummyEmail, DummyPassword)

		// Assert
		assert.Error(t, err)
	})
}

func TestAuth_SignOut(t *testing.T) {
	t.Run("closes database on sign out", func(t *testing.T) {
		// Arrange
		db := database.NewFake()
		client := api.NewFakeClient()
		auth := NewAuth(db, client, encryption.NewFakeManager())

		// Sign in
		_, err := auth.SignUp(DummyEmail, DummyPassword)
		require.NoError(t, err)

		// Act
		err = auth.SignOut()

		// Assert
		assert.NoError(t, err)
		assert.Panics(t, func() { _, _ = db.Read("test") }, "Read should panic if database connection was closed")
	})

	t.Run("panics when not signed in", func(t *testing.T) {
		// Arrange
		auth := Auth{}

		// Act/Assert
		assert.Panics(t, func() { _ = auth.SignOut() })
	})

	t.Run("returns error when database fails to close", func(t *testing.T) {
		// Arrange
		db := database.NewStub()
		db.CloseErr = errors.New("close error")
		auth := NewAuth(db, api.NewFakeClient(), encryption.NewFakeManager())
		auth.signedIn = true

		// Act
		err := auth.SignOut()

		// Assert
		assert.Error(t, err)
	})
}
