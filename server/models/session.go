package models

type Session struct {
	Account       Account            `json:"account"`
	Conversations []ConversationMeta `json:"conversations"`
}
