package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"reflect"
)

// Validate32ByteArray checks if the field is a [32]byte slice.
func Validate32ByteArray(fl validator.FieldLevel) bool {
	// Check if the field is a [32]byte slice
	return fl.Field().Kind() == reflect.Slice && fl.Field().Len() == 32
}

// Validate64ByteArray checks if the field is a [64]byte slice.
func Validate64ByteArray(fl validator.FieldLevel) bool {
	// Check if the field is a [64]byte slice
	return fl.Field().Kind() == reflect.Slice && fl.Field().Len() == 64
}

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("32bytes", Validate32ByteArray)
	_ = validate.RegisterValidation("64bytes", Validate64ByteArray)
	return &CustomValidator{validator: validate}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
