package auth

import (
	"encoding/base64"
	"errors"
	"github.com/labstack/echo/v4"
	"strings"
)

type Credentials struct {
	Username string
	Password string
}

//goland:noinspection GoErrorStringFormat
func ParseBasicAuthHeader(c echo.Context) (*Credentials, error) {
	// Extract Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// Basic auth format: "Basic base64(Username:Password)"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return nil, errors.New("invalid authorization header format")
	}

	// Decode the base64-encoded Username:Password
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid base64 encoding")
	}

	// Split Username and Password
	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return nil, errors.New("invalid username or password format")
	}

	return &Credentials{credentials[0], credentials[1]}, nil
}
