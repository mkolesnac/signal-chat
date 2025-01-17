package client

import (
	"encoding/json"
	"fmt"
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/stretchr/testify/assert"
	"net/http"
	"signal-chat/internal/api"
	"signal-chat/internal/client/database"
	"strings"
	"testing"
)

var dummyAPIClient = &APIClient{httpClient: &http.Client{Transport: &SpyRoundTripper{}}}

func TestSignUp(t *testing.T) {
	t.Run("returns error if email is invalid", func(t *testing.T) {
		auth := Auth{}
		err := auth.SignUp("", "password123")

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, ErrAuthInvalidEmail)
	})
	t.Run("returns error password shorter than 8 characters", func(t *testing.T) {
		auth := Auth{}
		err := auth.SignUp("test", "")

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, ErrAuthPwdTooShort)
	})
	t.Run("generates key pairs and writes them to db", func(t *testing.T) {
		db := database.NewFakeDatabase()
		auth := Auth{db: db, apiClient: dummyAPIClient}
		err := auth.SignUp("test", "password123")

		assert.Nil(t, err)
		assertWritesInDatabase(t, db, database.PrivateIdentityKeyPK(), 1)
		assertWritesInDatabase(t, db, database.PublicIdentityKeyPK(), 1)
		assertWritesInDatabase(t, db, database.SignedPreKeyPK(""), 1)
		assertWritesInDatabase(t, db, database.PreKeyPK(""), 100)
	})
	t.Run("sends public keys to server", func(t *testing.T) {
		db := database.NewFakeDatabase()
		spyTransport := &SpyRoundTripper{}
		apiClient := &APIClient{httpClient: &http.Client{Transport: spyTransport}}
		auth := Auth{db: db, apiClient: apiClient}
		err := auth.SignUp("test@user.com", "password123")

		assert.Nil(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.NotEmpty(t, spyTransport.Request.Header.Get("Authorization"), "authorization header should be set")

		var got api.SignUpRequest
		err = json.NewDecoder(spyTransport.Request.Body).Decode(&got)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "test@user.com", got.Email)
		assert.Equal(t, "password123", got.Password)
		assert.Len(t, got.IdentityPublicKey, 33, "curve25519 public keys should be 33 bytes long")
		assert.Equal(t, got.IdentityPublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		assert.Len(t, got.SignedPreKey.PublicKey, 33, "curve25519 public keys should be 33 bytes long")
		assert.Equal(t, got.SignedPreKey.PublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		assert.Len(t, got.SignedPreKey.Signature, 64, "Signed prekey signature should be 64 byte long")
		assert.Len(t, got.PreKeys, 100, "Request should contain 100 pre keys")
		for _, preKey := range got.PreKeys {
			assert.Len(t, preKey.PublicKey, 33, "curve25519 public keys should be 33 bytes long")
			assert.Equal(t, preKey.PublicKey[0], byte(ecc.DjbType), "curve25519 public keys should start with byte 0x05")
		}
	})
}

func TestSignIn(t *testing.T) {
	t.Run("returns error if email is invalid", func(t *testing.T) {
		auth := Auth{}
		err := auth.SignIn("", "password123")

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, ErrAuthInvalidEmail)
	})
	t.Run("returns error password shorter than 8 characters", func(t *testing.T) {
		auth := Auth{}
		err := auth.SignIn("test", "")

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, ErrAuthPwdTooShort)
	})
	t.Run("opens database connection", func(t *testing.T) {
		db := database.NewFakeDatabase()
		auth := Auth{db: db, apiClient: dummyAPIClient}
		err := auth.SignIn("test", "password123")

		assert.Nil(t, err)
		assert.NotPanics(t, func() { _, _ = db.ReadValue(database.PrivateIdentityKeyPK()) }, "Read should not panic if database connection was opened")
	})
	t.Run("sends signin request to server", func(t *testing.T) {
		db := database.NewFakeDatabase()
		spyTransport := &SpyRoundTripper{}
		apiClient := &APIClient{httpClient: &http.Client{Transport: spyTransport}}
		auth := Auth{db: db, apiClient: apiClient}
		err := auth.SignIn("test@user.com", "password123")

		assert.Nil(t, err)
		assert.NotNil(t, spyTransport.Request, "request should have been sent")
		assert.NotEmpty(t, spyTransport.Request.Header.Get("Authorization"), "authorization header should be set")

		var got api.SignInRequest
		err = json.NewDecoder(spyTransport.Request.Body).Decode(&got)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "test@user.com", got.Email)
		assert.Equal(t, "password123", got.Password)
	})
}

func TestSignOut(t *testing.T) {
	t.Run("panics if not signed in", func(t *testing.T) {
		auth := Auth{}
		assert.Panics(t, func() { _ = auth.SignOut() }, "Should panic if no user is signed in")
	})
	t.Run("closes database connection", func(t *testing.T) {
		db := database.NewFakeDatabase()
		auth := Auth{db: db, apiClient: dummyAPIClient}
		err := auth.SignIn("test@user.com", "password123")
		if err != nil {
			t.Fatal(err)
		}

		err = auth.SignOut()
		if err != nil {
			t.Fatal(err)
		}

		assert.Panics(t, func() { _, _ = db.ReadValue(database.PrivateIdentityKeyPK()) }, "Read should panic because database connection was closed")
	})
	t.Run("removes authentication from api client", func(t *testing.T) {
		db := database.NewFakeDatabase()
		spyTransport := &SpyRoundTripper{}
		apiClient := &APIClient{httpClient: &http.Client{Transport: spyTransport}}
		auth := Auth{db: db, apiClient: apiClient}
		err := auth.SignIn("test@user.com", "password123")
		if err != nil {
			t.Fatal(err)
		}

		err = auth.SignOut()
		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, apiClient.authorization)
	})
}

func assertWritesInDatabase(t *testing.T, db *database.FakeDatabase, prefix database.PrimaryKey, targetCount int) {
	t.Helper()

	count := 0
	for key, value := range db.Items {
		if strings.HasPrefix(string(key), string(prefix)) && value != nil && len(value) > 0 {
			count++
		}
	}

	if count != targetCount {
		keys := make([]string, 0, len(db.Items))
		for key := range db.Items {
			keys = append(keys, fmt.Sprintf("%q", key))
		}
		t.Errorf("Expected %d writes with key prefix %q in the database, but found %d. Database keys: [%s]",
			targetCount, prefix, count, strings.Join(keys, ", "))
	}
}
