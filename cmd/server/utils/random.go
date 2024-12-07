package utils

import "math/rand"

func RandomBytes(length int) []byte {
	byteArray := make([]byte, length)
	for i := range byteArray {
		byteArray[i] = byte(rand.Intn(256))
	}
	return byteArray
}
