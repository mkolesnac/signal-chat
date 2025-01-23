package apiclient

import (
	"errors"
	"net/http"
	"signal-chat/internal/api"
	"sync"
)

type User struct {
	username string
	password string
}

type Fake struct {
	users       map[string]User
	currentUser User
	mu          sync.RWMutex
}

func NewFake() *Fake {
	return &Fake{
		users: make(map[string]User),
	}
}

func (f *Fake) Get(route string, target any) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (f *Fake) Post(route string, payload any) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch route {
	case api.EndpointSignUp:
		req, ok := payload.(api.SignUpRequest)
		if !ok {
			return http.StatusBadRequest, errors.New("invalid payload")
		}

		if _, exists := f.users[req.UserName]; exists {
			return http.StatusBadRequest, errors.New("user already exists")
		}

		user := User{
			username: req.UserName,
			password: req.Password,
		}
		f.users[req.UserName] = user
		f.currentUser = user

		return http.StatusOK, nil
	case api.EndpointSignIn:
		req, ok := payload.(api.SignInRequest)
		if !ok {
			return http.StatusBadRequest, errors.New("invalid payload")
		}

		user, exists := f.users[req.Username]
		if !exists {
			return http.StatusBadRequest, errors.New("user not found")
		}
		if user.password != req.Password {
			return http.StatusBadRequest, errors.New("invalid password")
		}

		f.currentUser = user

		return http.StatusOK, nil
	default:
		return http.StatusNotFound, nil
	}
}
