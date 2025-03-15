package api

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignInResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Error    string `json:"error,omitempty"`
}
