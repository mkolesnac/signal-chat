package api

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/services"
)

// Upgrader configures the WebSocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin (only for development; specify origins in production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketHandler struct {
	websockets services.WebsocketManager
}

func NewWebsocketHandler(websockets services.WebsocketManager) *WebsocketHandler {
	return &WebsocketHandler{
		websockets: websockets,
	}
}

func (h *WebsocketHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/ws/:id", h.Upgrade)
}

func (h *WebsocketHandler) Upgrade(c echo.Context) error {
	accId := c.Param("id")

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade to WebSocket: %w", err)
	}
	h.websockets.RegisterClient(accId, conn)
	defer h.websockets.UnregisterClient(accId)

	return nil
}
