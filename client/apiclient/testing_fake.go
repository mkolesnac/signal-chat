package apiclient

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"reflect"
	"signal-chat/internal/api"
	"strconv"
	"sync"
	"time"
)

type User struct {
	id       string
	username string
	password string
}

type RequestRecord struct {
	Method      string
	Route       string
	Headers     map[string]string
	PayloadJSON []byte
}

type Fake struct {
	users         map[string]User
	currentUser   User
	authorization string
	mu            sync.RWMutex
	requests      []RequestRecord
}

func NewFake() *Fake {
	return &Fake{
		users: make(map[string]User),
	}
}

func (f *Fake) SetAuthorization(username, password string) {
	f.authorization = basicAuthorization(username, password)
}

func (f *Fake) ClearAuthorization() {
	f.authorization = ""
}

func (f *Fake) Get(route string, target any) (int, error) {
	f.recordRequest("GET", route, nil)

	//TODO implement me
	panic("implement me")
}

func (f *Fake) Post(route string, payload any, target any) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("POST", route, payload)

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
			id:       generateID(),
			username: req.UserName,
			password: req.Password,
		}
		f.users[req.UserName] = user
		f.currentUser = user

		response := api.SignUpResponse{
			UserID: user.id,
		}
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(response))
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

		response := api.SignInResponse{
			UserID: user.id,
		}
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(response))
		return http.StatusOK, nil
	default:
		return http.StatusNotFound, nil
	}
}

func (f *Fake) Requests() []RequestRecord {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return copy to prevent mutation
	result := make([]RequestRecord, len(f.requests))
	copy(result, f.requests)
	return result
}

func (f *Fake) recordRequest(method, route string, payload any) {
	r := RequestRecord{
		Method:  method,
		Route:   route,
		Headers: map[string]string{},
	}

	if payload != nil {
		payloadJSON, _ := json.Marshal(payload)
		r.PayloadJSON = payloadJSON
		r.Headers["Content-Type"] = "application/json"
	}

	if f.authorization != "" {
		r.Headers["Authorization"] = f.authorization
	}

	f.requests = append(f.requests, r)
}

func generateID() string {
	// Timestamp + 6 random digits
	return strconv.FormatInt(time.Now().UnixNano(), 10)[:13] + strconv.Itoa(rand.Intn(1000000))
}
