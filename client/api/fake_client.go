package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"signal-chat/internal/apitypes"
	"sync"
	"time"

	"github.com/google/uuid"
)

type user struct {
	id                string
	username          string
	password          string
	keyBundle         *apitypes.KeyBundle
	authToken         string
	pendingWSMessages []apitypes.WSMessage
}

type conversation struct {
	ID             string
	ParticipantIDs []string
}

type FakeClient struct {
	users       map[string]*user // by ID
	currentUser *user
	authTokens  map[string]string // authToken -> user ID

	mu                     sync.RWMutex
	username               string
	password               string
	conversations          map[string]conversation                   // all conversations
	handlers               map[apitypes.WSMessageType]MessageHandler // only for current user
	registrationIDs        map[string]uint32                         // for all users
	connectionStateHandler ConnectionStateHandler
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		users:           make(map[string]*user),
		authTokens:      make(map[string]string),
		handlers:        make(map[apitypes.WSMessageType]MessageHandler),
		conversations:   make(map[string]conversation),
		registrationIDs: make(map[string]uint32),
	}
}

func (f *FakeClient) SetWSMessageHandler(eventType apitypes.WSMessageType, handler MessageHandler) {
	if _, ok := f.handlers[eventType]; ok {
		panic(fmt.Sprintf("Handler for event %v already subscribed", eventType))
	}
	f.handlers[eventType] = handler
}

func (f *FakeClient) SetConnectionStateHandler(handler ConnectionStateHandler) {
	f.connectionStateHandler = handler
}

func (f *FakeClient) Close() {
	f.currentUser = nil
	f.handlers = make(map[apitypes.WSMessageType]MessageHandler)
}

func (f *FakeClient) SignUp(username, password string, keyBundle apitypes.KeyBundle) (apitypes.SignUpResponse, error) {
	userID := uuid.New().String()
	user := &user{
		id:        userID,
		username:  username,
		password:  password,
		keyBundle: &keyBundle,
	}

	f.users[userID] = user
	f.currentUser = user

	if f.connectionStateHandler != nil {
		f.connectionStateHandler(StateConnected)
	}

	return apitypes.SignUpResponse{
		UserID: userID,
	}, nil
}

func (f *FakeClient) SignIn(username, password string) (apitypes.SignInResponse, error) {
	for _, user := range f.users {
		if user.username == username && user.password == password {
			f.currentUser = user

			handler, handleExists := f.handlers[apitypes.MessageTypeSync]
			if handleExists && len(user.pendingWSMessages) > 0 {
				wsPayload := apitypes.WSSyncPayload{Messages: user.pendingWSMessages}
				handler(mustMarshal(wsPayload))
				user.pendingWSMessages = []apitypes.WSMessage{}
			}

			return apitypes.SignInResponse{
				UserID: user.id,
			}, nil
		}
	}

	if f.connectionStateHandler != nil {
		f.connectionStateHandler(StateConnected)
	}

	return apitypes.SignInResponse{}, &ServerError{
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid credentials",
	}
}

func (f *FakeClient) GetPreKeyBundle(id string) (apitypes.GetPreKeyBundleResponse, error) {
	user, exists := f.users[id]
	if !exists {
		return apitypes.GetPreKeyBundleResponse{}, &ServerError{StatusCode: http.StatusNotFound, Message: "User not found"}
	}

	if user.keyBundle == nil {
		return apitypes.GetPreKeyBundleResponse{}, &ServerError{StatusCode: http.StatusNotFound, Message: "User has no key bundle"}
	}

	registrationID := uint32(1234)
	if id, exists := f.registrationIDs[id]; exists {
		registrationID = id
	}

	// Get first preKey from the bundle
	if len(user.keyBundle.PreKeys) == 0 {
		return apitypes.GetPreKeyBundleResponse{}, &ServerError{StatusCode: http.StatusInternalServerError, Message: "No prekeys available"}
	}

	preKey := user.keyBundle.PreKeys[0]

	return apitypes.GetPreKeyBundleResponse{
		PreKeyBundle: apitypes.PreKeyBundle{
			RegistrationID: registrationID,
			IdentityKey:    user.keyBundle.IdentityKey,
			SignedPreKey:   user.keyBundle.SignedPreKey,
			PreKey:         preKey,
		},
	}, nil
}

func (f *FakeClient) GetUser(id string) (apitypes.GetUserResponse, error) {
	user, exists := f.users[id]
	if !exists {
		return apitypes.GetUserResponse{}, &ServerError{StatusCode: http.StatusNotFound, Message: "User not found"}
	}

	return apitypes.GetUserResponse{User: apitypes.User{
		ID:       user.id,
		Username: user.username,
	}}, nil
}

func (f *FakeClient) GetAllUsers() (apitypes.GetAllUsersResponse, error) {
	users := make([]apitypes.User, 0, len(f.users))
	for _, u := range f.users {
		users = append(users, apitypes.User{
			ID:       u.id,
			Username: u.username,
		})
	}

	return apitypes.GetAllUsersResponse{Users: users}, nil
}

func (f *FakeClient) CreateConversation(id string, otherParticipants []apitypes.Participant) error {
	if f.currentUser == nil {
		panic("This endpoint can only be used by authenticated user. Use SignUp or SignIn function for user authentication.")
	}

	participantIDs := []string{f.currentUser.id}

	for _, participant := range otherParticipants {
		participantIDs = append(participantIDs, participant.ID)

		// Prepare websocket message for the participant
		wsPayload := apitypes.WSNewConversationPayload{
			ConversationID:         id,
			SenderID:               f.currentUser.id,
			ParticipantIDs:         []string{f.currentUser.id},
			KeyDistributionMessage: participant.KeyDistributionMessage,
		}
		for _, wsParticipant := range otherParticipants {
			if wsParticipant.ID != participant.ID {
				wsPayload.ParticipantIDs = append(wsPayload.ParticipantIDs, wsParticipant.ID)
			}
		}

		wsMessage := apitypes.WSMessage{
			ID:   uuid.New().String(),
			Type: apitypes.MessageTypeNewConversation,
			Data: mustMarshal(wsPayload),
		}

		participant, exists := f.users[participant.ID]
		if !exists {
			panic(fmt.Sprintf("Participant %s is not registered in the api client", participant.id))
		}

		participant.pendingWSMessages = append(participant.pendingWSMessages, wsMessage)
	}

	conv := conversation{
		ID:             id,
		ParticipantIDs: participantIDs,
	}
	f.conversations[id] = conv

	return nil
}

func (f *FakeClient) SendMessage(conversationID string, content []byte) (apitypes.SendMessageResponse, error) {
	if f.currentUser == nil {
		panic("This endpoint can only be used by authenticated user. Use SignUp or SignIn function for user authentication.")
	}

	conversation := f.conversations[conversationID]
	msgID := uuid.New().String()
	timestamp := time.Now().UnixMilli()

	for _, id := range conversation.ParticipantIDs {
		if id != f.currentUser.id {
			wsPayload := apitypes.WSNewMessagePayload{
				ConversationID: conversationID,
				MessageID:      msgID,
				SenderID:       f.currentUser.id,
				Content:        content,
				CreatedAt:      timestamp,
			}
			wsMessage := apitypes.WSMessage{
				ID:   uuid.New().String(),
				Type: apitypes.MessageTypeNewMessage,
				Data: mustMarshal(wsPayload),
			}

			participant, exists := f.users[id]
			if !exists {
				panic(fmt.Sprintf("Participant %s is not registered in the api client", id))
			}

			participant.pendingWSMessages = append(participant.pendingWSMessages, wsMessage)
		}
	}

	return apitypes.SendMessageResponse{
		MessageID: msgID,
		CreatedAt: timestamp,
	}, nil
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal: %v", err))
	}
	return b
}
