package models

import "time"

func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
