package main

type PreKey struct {
	ID        uint32 `json:"id" validate:"required"`
	PublicKey []byte `json:"public_key" validate:"required,32bytes"`
}
