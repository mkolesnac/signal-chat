package services

import "errors"

var ErrAccountNotFound = errors.New("Account with the given ID not found")
var ErrInvalidSignature = errors.New("invalid signed pre-key signature")
var ErrConversationNotFound = errors.New("conversation with the given ID not found")
var ErrUnauthorized = errors.New("not authorized to access the specified resource")
var ErrNotParticipant = errors.New("the specified recipient is not a participant in the conversation")
