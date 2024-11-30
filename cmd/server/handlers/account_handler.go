package handlers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services"
	"strconv"
)

type AccountHandler struct {
	accounts services.AccountsService
}

func NewAccountHandler(accounts services.AccountsService) *AccountHandler {
	return &AccountHandler{
		accounts: accounts,
	}
}

func (h *AccountHandler) RegisterPublicRoutes(e *echo.Echo) {
	e.POST("/account", h.CreateAccount)
}

func (h *AccountHandler) RegisterPrivateRoutes(g *echo.Group) {
	g.GET("/account", h.GetCurrentSession)
	g.POST("/account/keys", h.UploadCurrentAccountKeys)
	g.GET("/account/keys/count", h.GetCurrentAccountKeyCount)
	g.GET("/account/:id", h.GetPublicAccountData)
}

func (h *AccountHandler) CreateAccount(c echo.Context) error {
	var req models.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	acc, err := h.accounts.CreateAccount(req.Name, req.Password, [32]byte(req.IdentityPublicKey), req.SignedPreKey, req.PreKeys)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSignature) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid signed prekey signature")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create account: %v", err))
		}
	}

	return c.JSON(http.StatusCreated, acc)
}

func (h *AccountHandler) GetCurrentSession(c echo.Context) error {
	acc := c.Get("account").(models.Account)
	res, err := h.accounts.GetSession(acc)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
	}
	return c.JSON(http.StatusOK, res)
}

func (h *AccountHandler) GetPublicAccountData(c echo.Context) error {
	id := c.Param("id")

	res, err := h.accounts.GetKeyBundle(id)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Target account not found")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get target account public keys")
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *AccountHandler) GetCurrentAccountKeyCount(c echo.Context) error {
	acc := c.Get("account").(models.Account)

	count, err := h.accounts.GetPreKeyCount(acc)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get pre key count")
	}

	return c.String(http.StatusOK, strconv.Itoa(count))
}

func (h *AccountHandler) UploadCurrentAccountKeys(c echo.Context) error {
	var req models.UploadPreKeysRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	acc := c.Get("account").(models.Account)

	// Upload signed prekey and one-time prekeys to storage
	err := h.accounts.UploadNewPreKeys(acc, req.SignedPreKey, req.PreKeys)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSignature) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid signed prekey signature")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to upload new pre keys: %v", err))
		}
	}

	return c.NoContent(http.StatusCreated)
}
