package models

type Conversation struct {
	Participants []Participant `json:"participants"`
	Messages     []Message     `json:"messages"`
}
