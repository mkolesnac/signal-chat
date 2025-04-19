package api

type SignUpRequest struct {
	UserName  string    `json:"username" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	KeyBundle KeyBundle `json:"keyBundle" validate:"required"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	UserID string `json:"userId"`
	Token  string `json:"token"`
}
