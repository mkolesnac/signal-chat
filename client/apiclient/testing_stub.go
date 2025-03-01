package apiclient

import (
	"net/http"
	"signal-chat/internal/api"
)

type StubResponse struct {
	StatusCode int
	Body       []byte
}

type Stub struct {
	StartSessionErr error
	CloseErr        error
	GetErrors       map[string]error
	PostErrors      map[string]error
	GetResponses    map[string]StubResponse
	PostResponses   map[string]StubResponse
	onWSError       api.WSErrorHandler
	handlers        map[api.WSMessageType]api.WSHandler
}

func NewStub() *Stub {
	return &Stub{
		GetErrors:     make(map[string]error),
		PostErrors:    make(map[string]error),
		GetResponses:  make(map[string]StubResponse),
		PostResponses: make(map[string]StubResponse),
		handlers:      make(map[api.WSMessageType]api.WSHandler),
	}
}

func (s *Stub) TriggerWebsocketMessages(messages []api.WSMessage) error {
	// Process each message and call the appropriate handler
	for _, msg := range messages {
		if handler, exists := s.handlers[msg.Type]; exists {
			if err := handler(msg.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Stub) StartSession(username, password string) error {
	return s.StartSessionErr
}

func (s *Stub) SetErrorHandler(handler api.WSErrorHandler) {
	s.onWSError = handler
}

func (s *Stub) Subscribe(eventType api.WSMessageType, handler api.WSHandler) {
	s.handlers[eventType] = handler
}

func (s *Stub) Get(route string) (int, []byte, error) {
	if err, ok := s.GetErrors[route]; ok {
		return 0, nil, err
	}
	if resp, ok := s.GetResponses[route]; ok {
		return resp.StatusCode, resp.Body, nil
	}

	return http.StatusOK, nil, nil
}

func (s *Stub) Post(route string, payload any) (int, []byte, error) {
	if err, ok := s.PostErrors[route]; ok {
		return 0, nil, err
	}
	if resp, ok := s.PostResponses[route]; ok {
		return resp.StatusCode, resp.Body, nil
	}

	return http.StatusOK, nil, nil
}

func (s *Stub) Close() error {
	return s.CloseErr
}
