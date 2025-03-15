package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"signal-chat/client/models"
	"signal-chat/internal/api"
)

var ErrAuthInvalidEmail = errors.New("email is not a valid email address")
var ErrAuthPwdTooShort = errors.New("password too short")

type AuthAPI interface {
	StartSession(username, password string) error
	Close() error
	Post(route string, payload any) (int, []byte, error)
}

type AuthDatabase interface {
	Open(userID string) error
	Close() error
}

type EncryptionInitializer interface {
	InitializeKeyStore() (api.KeyBundle, error)
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

	err := a.db.Open(email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to open user database: %w", err)
	}

	bundle, err := a.encryptor.InitializeKeyStore()
	if err != nil {
		return models.User{}, err
	}

	payload := api.SignUpRequest{
		UserName:  email,
		Password:  pwd,
		KeyBundle: bundle,
	}

	status, body, err := a.apiClient.Post(api.EndpointSignUp, payload)
	if err != nil {
		return models.User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return models.User{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.SignUpResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.User{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	err = a.apiClient.StartSession(email, pwd)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to start API session: %w", err)
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

	payload := api.SignInRequest{
		Username: email,
		Password: pwd,
	}

	status, body, err := a.apiClient.Post(api.EndpointSignIn, payload)
	if err != nil {
		return models.User{}, fmt.Errorf("got error from server: %w", err)
	}
	if status != http.StatusOK {
		return models.User{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.SignInResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.User{}, fmt.Errorf("got error unmarshalling response from server: %w", err)
	}

	err = a.apiClient.StartSession(email, pwd)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to start API session: %w", err)
	}

	a.signedIn = true
	user := models.User{
		ID:       resp.UserID,
		Username: resp.Username,
	}
	return user, nil
}

func (a *Auth) SignOut() error {
	if !a.signedIn {
		panic("not signed in")
	}

	if err := a.apiClient.Close(); err != nil {
		return fmt.Errorf("failed to close API session: %w", err)
	}

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
