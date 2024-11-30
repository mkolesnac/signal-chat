package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"signal-chat/cmd/server/auth"
	"signal-chat/cmd/server/handlers"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services"
	"signal-chat/cmd/server/storage"
)

func main() {
	backend := storage.NewMemoryStore()
	accounts := services.NewAccountsService(backend)
	keys := services.NewKeyService(backend, accounts)
	websockets := services.NewWebsocketManager()
	messages := services.NewMessageService(backend, accounts, websockets)

	accountsHandler := handlers.NewAccountHandler(accounts)
	keysHandler := handlers.NewKeysHandler(keys)
	messagesHandler := handlers.NewMessagesHandler(messages)
	websocketsHandler := handlers.NewWebsocketHandler(websockets)

	e := echo.New()
	e.Validator = models.NewCustomValidator()
	// Register public routes
	accountsHandler.RegisterPublicRoutes(e)

	// Register routes requiring authentication
	authGroup := e.Group("")
	authGroup.Use(middleware.BasicAuth(auth.BasicAuthMiddleware(accounts)))
	accountsHandler.RegisterPrivateRoutes(authGroup)
	keysHandler.RegisterRoutes(authGroup)
	messagesHandler.RegisterRoutes(authGroup)
	websocketsHandler.RegisterRoutes(authGroup)

	e.Logger.Fatal(e.Start(":8080"))
}
