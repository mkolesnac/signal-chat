package api

import (
	"signal-chat/internal/apitypes"
)

type StubClient struct {
	SignUpResponse          apitypes.SignUpResponse
	SignUpError             error
	SignInResponse          apitypes.SignInResponse
	SignInError             error
	GetPreKeyBundleResponse apitypes.GetPreKeyBundleResponse
	GetPreKeyBundleError    error
	GetUserResponse         apitypes.GetUserResponse
	GetUserError            error
	GetAllUsersResponse     apitypes.GetAllUsersResponse
	GetAllUsersError        error
	CreateConversationError error
	SendMessageResponse     apitypes.SendMessageResponse
	SendMessageError        error

	connectionStateHandler ConnectionStateHandler
	wsHandlers             map[apitypes.WSMessageType]MessageHandler
}

func NewStubClient() *StubClient {
	return &StubClient{
		wsHandlers: make(map[apitypes.WSMessageType]MessageHandler),
	}
}

func (s *StubClient) TriggerWebsocketMessages(messages []apitypes.WSMessage) {
	for _, msg := range messages {
		if handler, exists := s.wsHandlers[msg.Type]; exists {
			handler(msg.Data)
		}
	}
}

func (s *StubClient) SetWSMessageHandler(eventType apitypes.WSMessageType, handler MessageHandler) {
	s.wsHandlers[eventType] = handler
}

func (s *StubClient) SetConnectionStateHandler(handler ConnectionStateHandler) {
	s.connectionStateHandler = handler
}

func (s *StubClient) Close() {}

func (s *StubClient) SignUp(username, password string, keyBundle apitypes.KeyBundle) (apitypes.SignUpResponse, error) {
	if s.SignUpError != nil {
		return apitypes.SignUpResponse{}, s.SignUpError
	}

	return s.SignUpResponse, nil
}

func (s *StubClient) SignIn(username, password string) (apitypes.SignInResponse, error) {
	if s.SignInError != nil {
		return apitypes.SignInResponse{}, s.SignInError
	}

	return s.SignInResponse, nil
}

func (s *StubClient) GetPreKeyBundle(id string) (apitypes.GetPreKeyBundleResponse, error) {
	if s.GetPreKeyBundleError != nil {
		return apitypes.GetPreKeyBundleResponse{}, s.GetPreKeyBundleError
	}

	return s.GetPreKeyBundleResponse, nil
}

func (s *StubClient) GetUser(id string) (apitypes.GetUserResponse, error) {
	if s.GetUserError != nil {
		return apitypes.GetUserResponse{}, s.GetUserError
	}

	return s.GetUserResponse, nil
}

func (s *StubClient) GetAllUsers() (apitypes.GetAllUsersResponse, error) {
	if s.GetAllUsersError != nil {
		return apitypes.GetAllUsersResponse{}, s.GetAllUsersError
	}

	return s.GetAllUsersResponse, nil
}

func (s *StubClient) CreateConversation(id string, otherParticipants []apitypes.Participant) error {
	if s.CreateConversationError != nil {
		return s.CreateConversationError
	}

	return nil
}

func (s *StubClient) SendMessage(conversationID string, content []byte) (apitypes.SendMessageResponse, error) {
	if s.SendMessageError != nil {
		return apitypes.SendMessageResponse{}, s.SendMessageError
	}

	return s.SendMessageResponse, nil
}
