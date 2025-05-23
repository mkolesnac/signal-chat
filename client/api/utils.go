package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func takeRandomItem[T any](slice []T) (T, []T, error) {
	var result T

	if len(slice) == 0 {
		return result, slice, fmt.Errorf("cannot select from empty slice")
	}

	// Generate a secure random index
	maxBig := big.NewInt(int64(len(slice)))
	randomBig, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return result, slice, fmt.Errorf("failed to generate random number: %w", err)
	}

	randomIndex := int(randomBig.Int64())

	// get the selected item
	selectedItem := slice[randomIndex]

	// Remove the item from the slice
	newSlice := append(slice[:randomIndex], slice[randomIndex+1:]...)

	return selectedItem, newSlice, nil
}
