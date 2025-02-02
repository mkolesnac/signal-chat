package apiclient

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reflect"
	"signal-chat/internal/api"
	"strconv"
	"strings"
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
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("GET", route, nil)

	switch {
	case strings.HasPrefix(route, "/user/"):
		id := strings.TrimPrefix(route, "/user/")
		if id == "" {
			resp := api.GetUserResponse{Error: "invalid user ID"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}

		usr, exists := f.users[id]
		if !exists {
			resp := api.GetUserResponse{Error: "user not found"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}

		resp := api.GetUserResponse{
			Username: usr.username,
		}
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
		return http.StatusOK, nil
	//case route == "/user/":
	default:
		return http.StatusNotFound, nil
	}
}

func (f *Fake) Post(route string, payload any, target any) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("POST", route, payload)

	switch route {
	case api.EndpointSignUp:
		req, ok := payload.(api.SignUpRequest)
		if !ok {
			resp := api.SignUpResponse{Error: "invalid payload"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}

		if _, exists := f.users[req.UserName]; exists {
			resp := api.SignUpResponse{Error: "user already exists"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}

		user := User{
			id:       generateID(),
			username: req.UserName,
			password: req.Password,
		}
		f.users[req.UserName] = user
		f.currentUser = user

		resp := api.SignUpResponse{
			UserID: user.id,
		}
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
		return http.StatusOK, nil
	case api.EndpointSignIn:
		req, ok := payload.(api.SignInRequest)
		if !ok {
			resp := api.SignInResponse{Error: "invalid payload"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}

		user, exists := f.users[req.Username]
		if !exists {
			resp := api.SignInResponse{Error: "user not found"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
		}
		if user.password != req.Password {
			resp := api.SignInResponse{Error: "invalid password"}
			reflect.ValueOf(target).Elem().Set(reflect.ValueOf(resp))
			return http.StatusBadRequest, nil
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
