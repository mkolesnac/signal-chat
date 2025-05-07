package main

import (
	"errors"
	"fmt"
	"regexp"
	"signal-chat/client/models"
	"signal-chat/internal/apitypes"
)

var ErrAuthInvalidEmail = errors.New("email is not a valid email address")
var ErrAuthPwdTooShort = errors.New("password too short")

type AuthAPI interface {
	SignUp(username, password string, keyBundle apitypes.KeyBundle) (apitypes.SignUpResponse, error)
	SignIn(username, password string) (apitypes.SignInResponse, error)
	Close()
}

type AuthDatabase interface {
	Open(userID string) error
	Close() error
}

type EncryptionInitializer interface {
	InitializeKeyStore() (apitypes.KeyBundle, error)
}

type Auth struct {
	db        AuthDatabase
	apiClient AuthAPI
	encryptor EncryptionInitializer
	signedIn  bool
}

func NewAuth(db AuthDatabase, apiClient AuthAPI, encryptor EncryptionInitializer) *Auth {
	return &Auth{db: db, apiClient: apiClient, encryptor: encryptor}
}

func (a *Auth) SignUp(email, pwd string) (models.User, error) {
	if !isValidEmail(email) {
		return models.User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return models.User{}, ErrAuthPwdTooShort
	}

	if err := a.db.Open(email); err != nil {
		return models.User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	bundle, err := a.encryptor.InitializeKeyStore()
	if err != nil {
		return models.User{}, err
	}

	resp, err := a.apiClient.SignUp(email, pwd, bundle)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to sign up: %w", err)
	}

	a.signedIn = true
	user := models.User{
		ID:       resp.UserID,
		Username: email,
	}
	return user, nil
}

func (a *Auth) SignIn(email, pwd string) (models.User, error) {
	if !isValidEmail(email) {
		return models.User{}, ErrAuthInvalidEmail
	}
	if len(pwd) < 8 {
		return models.User{}, ErrAuthPwdTooShort
	}

	err := a.db.Open(email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	resp, err := a.apiClient.SignIn(email, pwd)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to sign in: %w", err)
	}

	a.signedIn = true
	user := models.User{
		ID:       resp.UserID,
		Username: email,
	}
	return user, nil
}

func (a *Auth) SignOut() error {
	if !a.signedIn {
		panic("not signed in")
	}

	a.apiClient.Close()
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

func isValidEmail(email string) bool {
	// Basic email regex
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}
