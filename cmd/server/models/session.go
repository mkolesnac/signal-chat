package models

type Session struct {
	Account       Account        `json:"account"`
	Conversations []Conversation `json:"conversations"`
}
