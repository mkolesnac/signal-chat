package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"signal-chat-server/api"
	"signal-chat-server/auth"
	"signal-chat-server/services"
	"signal-chat-server/storage"
)

func main() {
	backend := storage.NewMemoryStore()
	accounts := services.NewAccountService(backend)
	websockets := services.NewWebsocketManager()
	conversations := services.NewConversationService(backend, websockets)

	accountsHandler := api.NewAccountHandler(accounts)
	conversationHandler := api.NewConversationHandler(conversations)
	websocketsHandler := api.NewWebsocketHandler(websockets)

	e := echo.New()
	e.Validator = api.NewCustomValidator()
	// Register public routes
	accountsHandler.RegisterPublicRoutes(e)

	// Register routes requiring authentication
	authGroup := e.Group("")
	authGroup.Use(middleware.BasicAuth(auth.BasicAuthMiddleware(accounts)))
	accountsHandler.RegisterPrivateRoutes(authGroup)
	conversationHandler.RegisterRoutes(authGroup)
	websocketsHandler.RegisterRoutes(authGroup)

	e.Logger.Fatal(e.Start(":8080"))
}
