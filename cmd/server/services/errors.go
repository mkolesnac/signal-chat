package services

import "errors"

var ErrAccountNotFound = errors.New("account with the given ID not found")
var ErrInvalidSignature = errors.New("invalid signed pre-key signature")
var ErrConversationNotFound = errors.New("conversation with the given ID not found")
