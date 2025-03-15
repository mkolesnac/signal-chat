package main

import (
	"crypto/rand"
	"fmt"
)

func panicIfEmpty(argName, value string) {
	if len(value) == 0 {
		panic(fmt.Sprintf("%s cannot be empty", argName))
	}
}

// randomBytes generates a random byte array of the specified length.
// It uses crypto/rand for cryptographically secure random bytes.
func randomBytes(length int) []byte {
	if length < 0 {
		panic(fmt.Sprintf("invalid length: %d, length must be non-negative", length))
	}

	// Create a byte slice of the specified length
	bytes := make([]byte, length)

	// Fill the byte slice with random values
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}

	return bytes
}
