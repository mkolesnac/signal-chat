package auth

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"signal-chat/cmd/server/services"
)

func BasicAuthMiddleware(accounts services.AccountsService) func(username, password string, c echo.Context) (bool, error) {
	return func(username, password string, c echo.Context) (bool, error) {
		acc, err := accounts.GetAccount(username)
		if err != nil {
			if errors.Is(err, services.ErrAccountNotFound) {
				return false, echo.NewHTTPError(http.StatusBadRequest, "Account not found")
			} else {
				c.Logger().Errorf(err.Error())
				return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to get account")
			}
		}

		// Verify password
		err = bcrypt.CompareHashAndPassword(acc.PasswordHash, []byte(password))
		if err != nil {
			return false, nil
		}

		c.Set("account", acc)
		return true, nil
	}
}
