package api

type SignUpRequest struct {
	UserName  string    `json:"username" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	KeyBundle KeyBundle `json:"keyBundle" validate:"required"`
}

type SignUpResponse struct {
	UserID string `json:"userId"`
	Error  string `json:"error,omitempty"`
}
