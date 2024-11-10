package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"signal-chat/cmd/server/auth"
	"signal-chat/cmd/server/handlers"
	"signal-chat/cmd/server/services"
	"signal-chat/cmd/server/storage"
	"signal-chat/internal/api"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("32bytes", api.Validate32ByteArray)
	_ = validate.RegisterValidation("64bytes", api.Validate64ByteArray)
	return &CustomValidator{validator: validate}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	backend := storage.NewMemoryStorage()
	accounts := services.NewAccountService(backend)
	keys := services.NewKeyService(backend, accounts)
	websockets := services.NewWebsocketManager()
	messages := services.NewMessageService(backend, accounts, websockets)

	accountsHandler := handlers.NewAccountHandler(accounts)
	keysHandler := handlers.NewKeysHandler(keys)
	messagesHandler := handlers.NewMessagesHandler(messages)
	websocketsHandler := handlers.NewWebsocketHandler(websockets)

	e := echo.New()
	e.Validator = NewCustomValidator()
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
