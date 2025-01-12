package server

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"signal-chat/internal/server/handlers"
	"signal-chat/server/auth"
	"signal-chat/server/services"
	"signal-chat/server/storage"
)

type Options struct {
	Host string
	Port int
}

func Run(ctx context.Context, options *Options) {
	backend := storage.NewMemoryStore()
	accounts := services.NewAccountService(backend)
	websockets := services.NewWebsocketManager()
	conversations := services.NewConversationService(backend, websockets)

	accountsHandler := handlers.NewAccountHandler(accounts)
	conversationHandler := handlers.NewConversationHandler(conversations)
	websocketsHandler := handlers.NewWebsocketHandler(websockets)

	e := echo.New()
	e.Validator = handlers.NewCustomValidator()
	// Register public routes
	accountsHandler.RegisterPublicRoutes(e)

	// Register routes requiring authentication
	authGroup := e.Group("")
	authGroup.Use(middleware.BasicAuth(auth.BasicAuthMiddleware(accounts)))
	accountsHandler.RegisterPrivateRoutes(authGroup)
	conversationHandler.RegisterRoutes(authGroup)
	websocketsHandler.RegisterRoutes(authGroup)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%d", options.Host, options.Port)))
}
