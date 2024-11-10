package api

import (
	"encoding/base64"
	"github.com/go-playground/validator/v10"
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

func ValidateSignedPreKey(fl validator.FieldLevel) bool {
	// Get the string value of the field
	encodedStr := fl.Field().String()

	// Decode the base64 string
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return false
	}

	// Check if the decoded byte slice has exactly 64 bytes
	return len(decodedBytes) == 64
}
