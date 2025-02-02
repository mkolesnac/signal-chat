package main

import "time"

func isISO8601(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)
	return err == nil
}
