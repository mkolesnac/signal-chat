package handlers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	api2 "signal-chat/api"
	"signal-chat/server/models"
	"signal-chat/server/services"
	"strconv"
)

type AccountHandler struct {
	accounts services.AccountService
}

func NewAccountHandler(accounts services.AccountService) *AccountHandler {
	return &AccountHandler{
		accounts: accounts,
	}
}

func (h *AccountHandler) RegisterPublicRoutes(e *echo.Echo) {
	e.POST(api2.CreateUserEndpoint, h.CreateUser)
}

func (h *AccountHandler) RegisterPrivateRoutes(g *echo.Group) {
	g.POST("/account/keys", h.UploadCurrentAccountKeys)
	g.GET("/account/keys/count", h.GetCurrentAccountKeyCount)
	g.GET("/account/:id", h.GetAccountProfile)
	g.GET("/account/:id/keys", h.GetKeyBundle)
}

func (h *AccountHandler) CreateUser(c echo.Context) error {
	var req api2.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	acc, err := h.accounts.CreateAccount(req.Email, req.Password, [32]byte(req.IdentityPublicKey), req.SignedPreKey, req.PreKeys)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSignature) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid signed prekey signature")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create account: %v", err))
		}
	}

	res := api2.CreateUserResponse{ID: acc.ID}
	return c.JSON(http.StatusCreated, res)
}

func (h *AccountHandler) GetAccountProfile(c echo.Context) error {
	id := c.Param("id")

	res, err := h.accounts.GetAccount(id)
	if err != nil {
		if errors.Is(err, services.ErrAccountNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Target account not found")
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get target account public keys")
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *AccountHandler) GetKeyBundle(c echo.Context) error {
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
	var req api2.UploadPreKeysRequest
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
