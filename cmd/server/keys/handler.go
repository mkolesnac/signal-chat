package keys

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/accounts"
	"signal-chat/cmd/server/models"
	"strconv"
)

type Handler struct {
	accounts accounts.AccountService
	keys     KeyService
}

func NewHandler(keys KeyService) *Handler {
	return &Handler{
		keys: keys,
	}
}

// GetPreKeyCount get - Get prekey count
func (h *Handler) GetPreKeyCount(c echo.Context) error {
	acc := c.Get("account").(models.Account)

	count, err := h.keys.GetPreKeyCount(acc.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get pre key count")
	}

	return c.String(http.StatusOK, strconv.Itoa(count))
}

// uploadNewPreKeys - Uploads new prekeys
func (h *Handler) UploadNewPreKeys(c echo.Context) error {
	var req UploadPreKeysRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	acc := c.Get("account").(models.Account)

	// Verify signed prekey signature
	success, err := h.keys.VerifySignature(acc.ID, req.SignedPreKey.PublicKey, req.SignedPreKey.Signature)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify signature")
	}
	if !success {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid signature")
	}

	// Upload signed prekey and one-time prekeys to storage
	err = h.keys.UploadNewPreKeys(acc.ID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upload keys")
	}

	return c.NoContent(http.StatusCreated)
}

func (h *Handler) GetPublicKeys(c echo.Context) error {
	accId := c.QueryParam("id")

	resp, err := h.keys.GetPublicKeys(accId)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Account not found")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get public keys")
		}
	}

	return c.JSON(http.StatusOK, resp)
}
