package auth

import (
	"encoding/base64"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseAuthHeader(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		// Arrange
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderAuthorization, "Basic YWxhZGRpbjpvcGVuc2VzYW1l")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		credentials, _ := ParseBasicAuthHeader(c)

		// Assert
		assert.Equal(t, "aladdin", credentials.Username)
		assert.Equal(t, "opensesame", credentials.Password)
	})
	t.Run("missing authorization header", func(t *testing.T) {
		// Arrange
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		_, err := ParseBasicAuthHeader(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing authorization header")
	})
	t.Run("incorrect authorization scheme", func(t *testing.T) {
		// Arrange
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderAuthorization, "Digest username=\"aladdin\", realm=\"exampleRealm\", nonce=\"dcd98b7102dd2f0e8b11d0f600bfb0c093\", uri=\"/protected/resource\", response=\"94d80cdbb8f38665d381a156ea196a90\", qop=auth, nc=00000001, cnonce=\"a82c8f70f9c118b5\"")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		_, err := ParseBasicAuthHeader(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})
	t.Run("invalid base64 encoding", func(t *testing.T) {
		// Arrange
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderAuthorization, "Basic invalidbase64==")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		_, err := ParseBasicAuthHeader(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid base64 encoding")
	})
	t.Run("improper format", func(t *testing.T) {
		// Arrange
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderAuthorization, "Basic "+base64Encode("userpassword")) // Missing colon
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Act
		_, err := ParseBasicAuthHeader(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid username or password format")
	})
}

// Helper function to base64 encode a string
func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
