package storage

import "time"

func GetTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
