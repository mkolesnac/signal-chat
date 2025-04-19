package ws

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

func generateMessageID() string {
	b := make([]byte, 6) // 6 bytes = 48 bits of entropy
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	randomPart := base64.RawURLEncoding.EncodeToString(b)
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), randomPart)
}
