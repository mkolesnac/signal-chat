package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/client/database"
	"testing"
)

const DummyID = "123"

func TestUserService_GetUser(t *testing.T) {
	t.Run("fetches user data from server", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		auth := Auth{db: database.NewFake(), apiClient: ac}
		got, err := auth.SignUp("test@gmail.com", "test1234")
		require.NoError(t, err)
		svc := UserService{apiClient: ac}

		// Act
		want, err := svc.GetUser(got.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, got, want)
	})
	t.Run("panics when empty user ID", func(t *testing.T) {
		// Arrange
		svc := UserService{apiClient: apiclient.NewFake()}

		// Act & Assert
		assert.Panics(t, func() {
			_, _ = svc.GetUser("")
		})
	})
	t.Run("returns error when user doesn't exist", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewFake()
		svc := UserService{apiClient: ac}

		// Act
		_, err := svc.GetUser(DummyID)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when api client fails to send request", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.GetErr = errors.New("get error")
		svc := UserService{apiClient: ac}

		// Act
		_, err := svc.GetUser(DummyID)

		// Assert
		assert.Error(t, err)
	})
	t.Run("returns error when server returns unsuccessful response", func(t *testing.T) {
		// Arrange
		ac := apiclient.NewStub()
		ac.PostStatus = http.StatusInternalServerError
		svc := UserService{apiClient: ac}

		// Act
		_, err := svc.GetUser(DummyID)

		// Assert
		assert.Error(t, err)
	})
}
