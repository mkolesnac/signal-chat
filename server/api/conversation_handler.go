package api

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat-server/models"
	"signal-chat-server/services"
)

type ConversationHandler struct {
	conversations services.ConversationService
}

func NewConversationHandler(messages services.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		conversations: messages,
	}
}

func (h *ConversationHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/conversation", h.CreateConversation)
	g.GET("/conversation/:id", h.GetConversation)
	g.POST("/conversation/:id", h.SendMessage)
}

func (h *ConversationHandler) CreateConversation(c echo.Context) error {
	var req CreateConversationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	acc := c.Get("account").(models.Account)
	msg, err := h.conversations.CreateConversation(acc, req.CipherText, req.ParticipantIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create conversation")
	}

	return c.JSON(http.StatusOK, msg)
}

func (h *ConversationHandler) GetConversation(c echo.Context) error {
	id := c.Param("id")
	acc := c.Get("account").(models.Account)
	conversation, err := h.conversations.GetConversation(acc, id)
	if err != nil {
		if errors.Is(err, services.ErrConversationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Conversation not found")
		}
		if errors.Is(err, services.ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create conversation")
	}

	return c.JSON(http.StatusOK, conversation)

}

func (h *ConversationHandler) SendMessage(c echo.Context) error {
	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	id := c.Param("id")
	acc := c.Get("account").(models.Account)

	msg, err := h.conversations.SendMessage(acc, id, req.CipherText)
	if err != nil {
		if errors.Is(err, services.ErrConversationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Conversation not found")
		}
		if errors.Is(err, services.ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}
		if errors.Is(err, services.ErrNotParticipant) {
			return echo.NewHTTPError(http.StatusBadRequest, "Recipient is not a participant in the conversation")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send message")
	}

	return c.JSON(http.StatusOK, msg)
}
