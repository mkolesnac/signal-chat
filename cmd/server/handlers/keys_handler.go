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

type KeysHandler struct {
	keys services.KeyService
}

func NewKeysHandler(keys services.KeyService) *KeysHandler {
	return &KeysHandler{
		keys: keys,
	}
}

// GetPreKeyCount get - Get prekey count
func (h *KeysHandler) GetPreKeyCount(c echo.Context) error {
	acc := c.Get("account").(models.Account)

	count, err := h.keys.GetPreKeyCount(acc.GetID())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get pre key count")
	}

	return c.String(http.StatusOK, strconv.Itoa(count))
}

func (h *KeysHandler) UploadNewPreKeys(c echo.Context) error {
	var req api.UploadPreKeysRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	acc := c.Get("account").(models.Account)

	// Verify signed prekey signature
	success, err := h.keys.VerifySignature(acc.GetID(), req.SignedPreKey.PublicKey, req.SignedPreKey.Signature)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify signature")
	}
	if !success {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid signature")
	}

	// Upload signed prekey and one-time prekeys to storage
	err = h.keys.UploadNewPreKeys(acc.GetID(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upload keys")
	}

	return c.NoContent(http.StatusCreated)
}

func (h *KeysHandler) GetPublicKeys(c echo.Context) error {
	accId := c.QueryParam("id")

	resp, err := h.keys.GetPublicKeys(accId)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Target account not found")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get public keys")
		}
	}

	return c.JSON(http.StatusOK, resp)
}
