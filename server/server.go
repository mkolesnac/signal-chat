package main

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"signal-chat/internal/apitypes"
	"signal-chat/server/conversation"
	"signal-chat/server/ws"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow clients from any origin (only for development; specify origins in production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Authenticator interface {
	GenerateToken(userID string) (string, error)
	Authenticate(r *http.Request) (string, error)
	RevokeToken(r *http.Request) error
}

type WebsocketManager interface {
	RegisterClient(clientID string, conn ws.Connection) error
	UnregisterClient(clientID string)
	BroadcastNewConversation(senderID string, req apitypes.CreateConversationRequest) error
	BroadcastNewMessage(senderID, messageID string, req apitypes.SendMessageRequest) error
}

type Server struct {
	router            *echo.Echo
	userStore         *UserStore
	conversationStore *conversation.Store
	auth              Authenticator
	wsManager         WebsocketManager
}

type ServerConfig struct {
	ReadTimeout  int
	WriteTimeout int
	MaxBodySize  string
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ReadTimeout:  60,
		WriteTimeout: 60,
		MaxBodySize:  "10MB",
	}
}

func NewServer(db *badger.DB) (*Server, error) {
	return NewServerWithConfig(db, DefaultServerConfig())
}

func NewServerWithConfig(db *badger.DB, config ServerConfig) (*Server, error) {
	e := echo.New()

	// Configure server timeouts and limits
	e.Server.ReadTimeout = time.Duration(config.ReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(config.WriteTimeout) * time.Second
	e.Server.MaxHeaderBytes = 1 << 20 // 1MB

	// Set custom validator
	e.Validator = NewCustomValidator()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit(config.MaxBodySize))

	convStore := conversation.NewStore(db)

	server := &Server{
		router:            e,
		userStore:         &UserStore{db: db},
		conversationStore: convStore,
		auth:              NewAuthManager(),
		wsManager:         ws.NewManager(db, convStore),
	}

	// Register routes
	e.GET(apitypes.EndpointUser, server.handleGetUser)
	e.GET(apitypes.EndpointPreKeyBundle, server.handleGetUserKeys)
	e.GET(apitypes.EndpointUsers, server.handleGetAllUsers)

	e.POST(apitypes.EndpointSignUp, server.handleSignUp)
	e.POST(apitypes.EndpointSignIn, server.handleSignIn)
	e.POST(apitypes.EndpointSignOut, server.handleSignOut)
	e.POST(apitypes.EndpointConversations, server.handleCreateConversation)
	e.POST(apitypes.EndpointMessages, server.handleCreateMessage)

	// Add WebSocket endpoint
	e.GET("/ws", server.handleWebSocketConnection)

	return server, nil
}

// Start starts the server on the specified host and port
func (s *Server) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Starting server on %s", addr)
	return s.router.Start(addr)
}

func (s *Server) handleSignUp(c echo.Context) error {
	var req apitypes.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	usr, err := s.userStore.CreateUser(req.Username, req.Password, req.KeyBundle)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return echo.NewHTTPError(http.StatusConflict, "failed to create new user")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create new user")
	}

	token, err := s.auth.GenerateToken(usr.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate user token")
	}

	resp := apitypes.SignUpResponse{
		UserID:    usr.ID,
		AuthToken: token,
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleSignIn(c echo.Context) error {
	var req apitypes.SignInRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	usr, err := s.userStore.VerifyCredentials(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid username or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify credentials")
	}

	token, err := s.auth.GenerateToken(usr.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate user token")
	}

	resp := apitypes.SignInResponse{
		UserID:    usr.ID,
		AuthToken: token,
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleSignOut(c echo.Context) error {
	err := s.auth.RevokeToken(c.Request())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to sign out")
	}
	return echo.NewHTTPError(http.StatusOK)
}

func (s *Server) handleGetUser(c echo.Context) error {
	if _, err := s.authenticate(c); err != nil {
		return err
	}

	id := c.Param("id")

	user, err := s.userStore.GetUserByID(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, apitypes.GetUserResponse{User: user})
}

func (s *Server) handleGetAllUsers(c echo.Context) error {
	if _, err := s.authenticate(c); err != nil {
		return err
	}

	users, err := s.userStore.GetAllUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, apitypes.GetAllUsersResponse{Users: users})
}

func (s *Server) handleGetUserKeys(c echo.Context) error {
	if _, err := s.authenticate(c); err != nil {
		return err
	}

	id := c.Param("id")
	bundle, err := s.userStore.GetPreKeyBundle(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, apitypes.GetPreKeyBundleResponse{PreKeyBundle: bundle})
}

func (s *Server) handleCreateConversation(c echo.Context) error {
	userID, authErr := s.authenticate(c)
	if authErr != nil {
		return authErr
	}

	var req apitypes.CreateConversationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	participantIDs := make([]string, 0, len(req.OtherParticipants)+1)
	participantIDs = append(participantIDs, userID)
	for _, r := range req.OtherParticipants {
		participantIDs = append(participantIDs, r.ID)
	}

	err := s.conversationStore.CreateConversation(req.ConversationID, participantIDs)
	if err != nil {
		if errors.Is(err, conversation.ErrConversationExists) {
			return echo.NewHTTPError(http.StatusConflict, "failed to create conversation")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create conversation")
	}

	// Broadcast the new conversation to all participants
	if err := s.wsManager.BroadcastNewConversation(userID, req); err != nil {
		log.Printf("Failed to broadcast new conversation: %v", err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) handleCreateMessage(c echo.Context) error {
	userID, authErr := s.authenticate(c)
	if authErr != nil {
		return authErr
	}

	var req apitypes.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	messageID, err := s.conversationStore.CreateMessage(userID, req.ConversationID, req.Content)
	if err != nil {
		if errors.Is(err, conversation.ErrConversationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		} else if errors.Is(err, conversation.ErrConversationUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create message")
	}

	// Broadcast the new message to all participants
	if err := s.wsManager.BroadcastNewMessage(userID, messageID, req); err != nil {
		log.Printf("Failed to broadcast new message: %v", err)
		// Continue even if broadcasting fails
	}

	return c.JSON(http.StatusOK, apitypes.SendMessageResponse{MessageID: messageID})
}

func (s *Server) handleWebSocketConnection(c echo.Context) error {
	userID, authErr := s.authenticate(c)
	if authErr != nil {
		return authErr
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}

	err = s.wsManager.RegisterClient(userID, conn)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register websocket listener")
	}

	// Set up a cleanup function to unregister the client when the connection is closed
	go func() {
		// Wait for the connection to be closed
		<-c.Request().Context().Done()
		s.wsManager.UnregisterClient(userID)
	}()

	return nil
}

func (s *Server) authenticate(c echo.Context) (string, *echo.HTTPError) {
	userID, err := s.auth.Authenticate(c.Request())

	if err != nil {
		switch {
		case errors.Is(err, ErrTokenUnauthorized):
			return "", echo.NewHTTPError(http.StatusUnauthorized, "unauthorized token")
		case errors.Is(err, ErrMissingAuthHeader),
			errors.Is(err, ErrEmptyToken),
			errors.Is(err, ErrDecodeToken):
			return "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
		default:
			return "", echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}

	return userID, nil
}
