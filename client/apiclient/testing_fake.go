package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"signal-chat/client/models"
	"signal-chat/internal/api"
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
	mu            sync.RWMutex
	requests      []RequestRecord
	username      string
	password      string
	conversations map[string]models.Conversation      // all conversations
	userSyncData  map[string]api.WSSyncData           // for all users
	handlers      map[api.WSMessageType]api.WSHandler // only for current user
	onWSError     api.WSErrorHandler
	RequireAuth   bool
}

func NewFake() *Fake {
	return &Fake{
		users:         make(map[string]User),
		handlers:      make(map[api.WSMessageType]api.WSHandler),
		conversations: make(map[string]models.Conversation),
		userSyncData:  map[string]api.WSSyncData{},
	}
}

func NewFakeWithoutAuth() *Fake {
	f := Fake{
		users:         make(map[string]User),
		handlers:      make(map[api.WSMessageType]api.WSHandler),
		conversations: make(map[string]models.Conversation),
		userSyncData:  map[string]api.WSSyncData{},
	}

	id := uuid.New().String()
	f.users[id] = User{
		id:       id,
		username: "Dummy",
		password: "Dummy",
	}
	f.username = "Dummy"
	f.password = "Dummy"
	return &f
}

func (f *Fake) StartSession(username, password string) error {
	if len(f.handlers) > 0 {
		panic("Close connection of the previous user before starting a new session")
	}
	f.username = username
	f.password = password
	_, err := f.authenticate(f.username, f.password)
	if err != nil {
		return err
	}

	syncData, dataExists := f.userSyncData[username]
	handler, handleExists := f.handlers[api.MessageTypeSync]
	if dataExists && handleExists {
		jsonData, err := json.Marshal(syncData)
		if err != nil {
			return err
		}

		err = handler(jsonData)
		if err != nil {
			f.handleWSError(err)
		}
	}

	return nil
}

func (f *Fake) SetErrorHandler(handler api.WSErrorHandler) {
	f.onWSError = handler
}

func (f *Fake) Close() error {
	f.username = ""
	f.password = ""
	f.handlers = make(map[api.WSMessageType]api.WSHandler)
	return nil
}

func (f *Fake) Subscribe(eventType api.WSMessageType, handler api.WSHandler) {
	f.handlers[eventType] = handler
}

func (f *Fake) Get(route string) (int, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("GET", route, nil)

	switch {
	case strings.HasPrefix(route, api.EndpointUser):
		id := strings.TrimPrefix(strings.TrimPrefix(route, api.EndpointUser), "/")
		if id == "" {
			return f.badRequestResponse("invalid user ID")
		}

		usr, exists := f.users[id]
		if !exists {
			return f.badRequestResponse("user not found")
		}

		return f.respond(http.StatusOK, api.GetUserResponse{Username: usr.username})
	//case route == "/user/":
	default:
		return http.StatusNotFound, nil, nil
	}
}

func (f *Fake) Post(route string, payload any) (int, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("POST", route, payload)

	var sender User
	var err error
	if route != api.EndpointSignUp && route != api.EndpointSignIn {
		sender, err = f.authenticate(f.username, f.password)
		if err != nil {
			return http.StatusUnauthorized, nil, nil
		}
	}

	switch route {
	case api.EndpointSignUp:
		req, ok := payload.(api.SignUpRequest)
		if !ok {
			return f.badRequestResponse("invalid payload")
		}

		if _, exists := f.users[req.UserName]; exists {
			return f.badRequestResponse("sender already exists")
		}

		user := User{
			id:       uuid.New().String(),
			username: req.UserName,
			password: req.Password,
		}
		f.users[user.id] = user
		return f.respond(http.StatusOK, api.SignUpResponse{UserID: user.id})
	case api.EndpointSignIn:
		req, ok := payload.(api.SignInRequest)
		if !ok {
			return f.badRequestResponse("invalid payload")
		}

		usr, err := f.authenticate(req.Username, req.Password)
		if err != nil {
			return http.StatusUnauthorized, nil, nil
		}

		return f.respond(http.StatusOK, api.SignInResponse{
			UserID:   usr.id,
			Username: usr.username,
		})
	case api.EndpointConversations:
		req, ok := payload.(api.CreateConversationRequest)
		if !ok {
			return f.badRequestResponse("invalid payload")
		}

		convID := uuid.New().String()
		msgID := uuid.New().String()
		timestamp := time.Now().Format(time.RFC3339)
		participantIDs := append(req.RecipientIDs, sender.id)
		syncData := f.userSyncData[sender.id]
		syncData.NewConversations = append(syncData.NewConversations, api.WSNewConversationPayload{
			ConversationID: convID,
			ParticipantIDs: participantIDs,
			SenderID:       sender.id,
			MessageID:      msgID,
			MessageText:    req.MessageText,
			MessagePreview: req.MessagePreview,
			Timestamp:      timestamp,
		})
		f.userSyncData[sender.id] = syncData

		conv := models.Conversation{
			ID:                   convID,
			LastMessagePreview:   req.MessagePreview,
			LastMessageSenderID:  sender.id,
			LastMessageTimestamp: timestamp,
			ParticipantIDs:       participantIDs,
		}
		f.conversations[convID] = conv

		return f.respond(http.StatusOK, api.CreateConversationResponse{
			ConversationID: convID,
			MessageID:      msgID,
			ParticipantIDs: participantIDs,
			Timestamp:      timestamp,
		})
	case api.EndpointMessages:
		req, ok := payload.(api.CreateMessageRequest)
		if !ok {
			return f.badRequestResponse("invalid payload")
		}

		conv, exists := f.conversations[req.ConversationID]
		if !exists {
			return f.badRequestResponse("conversation not found")
		}

		msgID := uuid.New().String()
		timestamp := time.Now().Format(time.RFC3339)

		for _, id := range conv.ParticipantIDs {
			if id != sender.id {
				syncData := f.userSyncData[sender.id]
				syncData.NewMessages = append(syncData.NewMessages, api.WSNewMessagePayload{
					ConversationID: conv.ID,
					MessageID:      msgID,
					SenderID:       sender.id,
					Text:           req.Text,
					Preview:        req.Preview,
					Timestamp:      timestamp,
				})
				f.userSyncData[sender.id] = syncData
			}
		}

		return f.respond(http.StatusOK, api.CreateMessageResponse{
			MessageID: msgID,
			SenderID:  sender.id,
			Timestamp: timestamp,
		})
	}

	return http.StatusNotFound, nil, nil
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

	if f.username != "" {
		r.Headers["Authorization"] = basicAuthorization(f.username, f.password)
	}

	f.requests = append(f.requests, r)
}

func (f *Fake) authenticate(username, password string) (User, error) {
	for _, u := range f.users {
		if u.username == username && u.password == password {
			return u, nil
		}
	}
	return User{}, errors.New("no user with matching credentials was found")
}

func (f *Fake) respond(status int, response interface{}) (int, []byte, error) {
	b, err := json.Marshal(response)
	if err != nil {
		return 0, nil, err
	}
	return status, b, nil
}

// Helper for error responses
func (f *Fake) badRequestResponse(msg string) (int, []byte, error) {
	return http.StatusBadRequest, []byte(fmt.Sprintf(`{"error":"%s"}`, msg)), nil
}

func (f *Fake) handleWSError(err error) {
	if f.onWSError != nil {
		f.onWSError(err)
	} else {
		panic(fmt.Errorf("unhlandled websocket error: %w", err))
	}
}
