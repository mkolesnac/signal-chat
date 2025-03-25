package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"signal-chat/internal/api"
	"strings"
	"sync"
	"time"
)

type User struct {
	id        string
	username  string
	password  string
	keyBundle *api.KeyBundle
}

type Conversation struct {
	ID             string
	Name           string
	ParticipantIDs []string
}

type RequestRecord struct {
	Method      string
	Route       string
	Headers     map[string]string
	PayloadJSON []byte
}

type Fake struct {
	users           map[string]User
	mu              sync.RWMutex
	requests        []RequestRecord
	username        string
	password        string
	conversations   map[string]Conversation             // all conversations
	userSyncData    map[string]api.WSSyncData           // for all users
	handlers        map[api.WSMessageType]api.WSHandler // only for current user
	registrationIDs map[string]uint32                   // for all users
	onWSError       api.WSErrorHandler
	RequireAuth     bool
}

func NewFake() *Fake {
	return &Fake{
		users:           make(map[string]User),
		handlers:        make(map[api.WSMessageType]api.WSHandler),
		conversations:   make(map[string]Conversation),
		userSyncData:    make(map[string]api.WSSyncData),
		registrationIDs: make(map[string]uint32),
	}
}

func NewFakeWithoutAuth() *Fake {
	f := Fake{
		users:           make(map[string]User),
		handlers:        make(map[api.WSMessageType]api.WSHandler),
		conversations:   make(map[string]Conversation),
		userSyncData:    make(map[string]api.WSSyncData),
		registrationIDs: make(map[string]uint32),
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
	f.username = username
	f.password = password
	usr, err := f.authenticate(f.username, f.password)
	if err != nil {
		return err
	}

	syncData, dataExists := f.userSyncData[usr.id]
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
	if _, ok := f.handlers[eventType]; ok {
		panic(fmt.Sprintf("Handler for event %v already subscribed", eventType))
	}
	f.handlers[eventType] = handler
}

func (f *Fake) ClearHandlers() {
	f.handlers = make(map[api.WSMessageType]api.WSHandler)
}

func (f *Fake) Get(route string) (int, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recordRequest("GET", route, nil)

	switch {
	case matchRoutePattern(route, api.EndpointUser(":id")):
		userID := extractParam(route, api.EndpointUserKeys(":id"), "id")
		usr, exists := f.users[userID]
		if !exists {
			return f.badRequestResponse("user not found")
		}

		return f.respond(http.StatusOK, api.GetUserResponse{Username: usr.username})
	case matchRoutePattern(route, api.EndpointUserKeys(":id")):
		userID := extractParam(route, api.EndpointUserKeys(":id"), "id")
		usr, exists := f.users[userID]
		if !exists {
			return f.badRequestResponse("user not found")
		}

		preKey, newPreKeys, err := takeRandomItem(usr.keyBundle.PreKeys)
		if err != nil {
			return f.internalErrorResponse("failed to select pre key")
		}
		usr.keyBundle.PreKeys = newPreKeys

		bundle := api.PreKeyBundle{
			RegistrationID: usr.keyBundle.RegistrationID,
			IdentityKey:    usr.keyBundle.IdentityKey,
			SignedPreKey: api.PreKey{
				ID:        usr.keyBundle.SignedPreKey.ID,
				PublicKey: usr.keyBundle.SignedPreKey.PublicKey,
			},
			SignedPreKeySignature: usr.keyBundle.SignedPreKey.Signature,
			PreKey:                preKey,
		}
		return f.respond(http.StatusOK, api.GetPrekeyBundleResponse{PreKeyBundle: bundle})
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
			//id:        uuid.New().String(),
			id:        req.UserName,
			username:  req.UserName,
			password:  req.Password,
			keyBundle: &req.KeyBundle,
		}
		f.users[user.id] = user
		f.registrationIDs[user.id] = rand.Uint32()

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

		for _, current := range req.Recipients {
			otherRecipientIDs := []string{sender.id}
			for _, r := range req.Recipients {
				if r.ID != current.ID {
					otherRecipientIDs = append(otherRecipientIDs, r.ID)
				}
			}

			syncData := f.userSyncData[current.ID]
			syncData.NewConversations = append(syncData.NewConversations, api.WSNewConversationPayload{
				ConversationID:         req.ConversationID,
				SenderID:               sender.id,
				RecipientIDs:           otherRecipientIDs,
				KeyDistributionMessage: current.KeyDistributionMessage,
			})
			f.userSyncData[current.ID] = syncData
		}

		recipientIDs := []string{sender.id}
		for _, r := range req.Recipients {
			recipientIDs = append(recipientIDs, r.ID)
		}
		conv := Conversation{
			ID:             req.ConversationID,
			ParticipantIDs: recipientIDs,
		}
		f.conversations[conv.ID] = conv

		return f.respond(http.StatusOK, api.CreateConversationResponse{})
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
		timestamp := time.Now().UnixMilli()

		for _, id := range conv.ParticipantIDs {
			if id != sender.id {
				syncData := f.userSyncData[id]
				syncData.NewMessages = append(syncData.NewMessages, api.WSNewMessagePayload{
					ConversationID:   conv.ID,
					MessageID:        msgID,
					SenderID:         sender.id,
					EncryptedMessage: req.EncryptedMessage,
					Timestamp:        timestamp,
				})
				f.userSyncData[id] = syncData
			}
		}

		return f.respond(http.StatusOK, api.CreateMessageResponse{
			MessageID: msgID,
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

func (f *Fake) internalErrorResponse(msg string) (int, []byte, error) {
	return http.StatusInternalServerError, []byte(fmt.Sprintf(`{"error":"%s"}`, msg)), nil
}

func (f *Fake) handleWSError(err error) {
	if f.onWSError != nil {
		f.onWSError(err)
	} else {
		panic(fmt.Errorf("unhlandled websocket error: %w", err))
	}
}

// Helper functions to match route patterns and extract parameters
func matchRoutePattern(actualRoute, patternRoute string) bool {
	// Split both routes into segments
	actualSegments := strings.Split(strings.Trim(actualRoute, "/"), "/")
	patternSegments := strings.Split(strings.Trim(patternRoute, "/"), "/")

	// If segment counts don't match, routes don't match
	if len(actualSegments) != len(patternSegments) {
		return false
	}

	// Compare each segment
	for i, patternSeg := range patternSegments {
		// If this is a parameter segment (starts with :), it matches anything
		if strings.HasPrefix(patternSeg, ":") {
			continue
		}

		// Otherwise segments must match exactly
		if patternSeg != actualSegments[i] {
			return false
		}
	}

	return true
}

func extractParam(actualRoute, patternRoute, paramName string) string {
	// Split both routes into segments
	actualSegments := strings.Split(strings.Trim(actualRoute, "/"), "/")
	patternSegments := strings.Split(strings.Trim(patternRoute, "/"), "/")

	// Find the index of the parameter
	paramIndex := -1
	for i, segment := range patternSegments {
		if segment == ":"+paramName {
			paramIndex = i
			break
		}
	}

	// If parameter not found or out of bounds in actual route
	if paramIndex == -1 || paramIndex >= len(actualSegments) {
		return ""
	}

	return actualSegments[paramIndex]
}
