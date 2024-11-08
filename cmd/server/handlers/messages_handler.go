package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services"
	"signal-chat/internal/api"
	"strconv"
)

type MessagesHandler struct {
	messages services.MessageService
}

func NewMessagesHandler(messages services.MessageService) *MessagesHandler {
	return &MessagesHandler{
		messages: messages,
	}
}

func (h *MessagesHandler) GetMessages(c echo.Context) error {
	acc := c.Get("account").(models.Account)
	fromS := c.QueryParam("from")
	from, err := strconv.ParseInt(fromS, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid from parameter")
	}

	messages, err := h.messages.GetMessages(acc.GetID(), from)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get messages: "+err.Error())
	}
	var respMessages []api.Message
	for _, m := range messages {
		respMessages = append(respMessages, api.Message{
			ID:         m.GetID(),
			SenderID:   m.SenderID,
			CipherText: m.CipherText,
			CreatedAt:  m.CreatedAt,
		})
	}

	resp := api.GetMessagesResponse{Messages: respMessages}
	return c.JSON(http.StatusOK, resp)
}

func (h *MessagesHandler) SendMessage(c echo.Context) error {
	var req api.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {

		return err
	}

	acc := c.Get("account").(models.Account)
	recipientID := c.QueryParam("recipient")

	id, err := h.messages.SendMessage(acc.GetID(), recipientID, req)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Recipient account not found")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get public keys")
		}
	}

	resp := api.SendMessageResponse{MessageID: id}
	return c.JSON(http.StatusOK, resp)
}
