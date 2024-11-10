package handlers

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/models"
	"signal-chat/cmd/server/services"
	"signal-chat/internal/api"
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
	e.POST("/account", h.CreateAccount)
}

func (h *AccountHandler) RegisterPrivateRoutes(g *echo.Group) {
	g.GET("/account", h.GetAccount)
}

func (h *AccountHandler) CreateAccount(c echo.Context) error {
	var req api.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// Validate signature of the signed pre-key
	signedKeyValid := ecc.VerifySignature(ecc.NewDjbECPublicKey([32]byte(req.IdentityPublicKey)), req.SignedPreKey.PublicKey, [64]byte(req.SignedPreKey.Signature))
	if !signedKeyValid {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid signed prekey signature")
	}

	id, err := h.accounts.CreateAccount(req.Name, req.Password, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create account")
	}

	res := api.CreateAccountResponse{ID: id}
	return c.JSON(http.StatusCreated, res)
}

func (h *AccountHandler) GetAccount(c echo.Context) error {
	acc := c.Get("account").(models.Account)
	res := api.GetAccountResponse{Name: acc.Name, CreatedAt: acc.CreatedAt}
	return c.JSON(http.StatusOK, res)
}
