package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrTokenUnauthorized = errors.New("invalid or revoked token")
	ErrMissingAuthHeader = errors.New("missing or invalid auth scheme")
	ErrEmptyToken        = errors.New("empty token")
	ErrDecodeToken       = errors.New("failed to decode token")
)

type AuthManager struct {
	tokens sync.Map
}

func NewAuthManager() *AuthManager {
	return &AuthManager{tokens: sync.Map{}}
}

func (m *AuthManager) GenerateToken(userID string) (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	hash := hashToken(token)
	m.tokens.Store(hash, userID)

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func (m *AuthManager) Authenticate(r *http.Request) (string, error) {
	token, err := getToken(r)
	if err != nil {
		return "", err
	}

	hash := hashToken(token)
	if userID, ok := m.tokens.Load(hash); ok {
		return userID.(string), nil
	}

	return "", ErrTokenUnauthorized
}

func (m *AuthManager) RevokeToken(r *http.Request) error {
	token, err := getToken(r)
	if err != nil {
		return err
	}

	hash := hashToken(token)
	m.tokens.Delete(hash)

	return nil
}

func getToken(r *http.Request) ([]byte, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, ErrMissingAuthHeader
	}

	encoded := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if encoded == "" {
		return nil, ErrEmptyToken
	}

	token, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeToken, err)
	}

	return token, nil
}

func hashToken(token []byte) string {
	sum := sha256.Sum256(token)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
