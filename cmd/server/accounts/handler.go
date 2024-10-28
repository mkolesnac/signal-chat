package accounts

import (
	"github.com/crossle/libsignal-protocol-go/ecc"
	"github.com/labstack/echo/v4"
	"net/http"
	"signal-chat/cmd/server/auth"
)

type Handler struct {
	accounts AccountService
}

func NewHandler(accounts AccountService) *Handler {
	return &Handler{
		accounts: accounts,
	}
}

// PUT /account
func (h *Handler) CreateAccount(c echo.Context) error {
	var req CreateAccountRequest
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

	credentials, err := auth.ParseBasicAuthHeader(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.accounts.CreateAccount(credentials.Username, credentials.Password, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create account")
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Account created successfully"})
}