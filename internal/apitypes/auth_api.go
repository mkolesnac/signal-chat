package apitypes

type SignUpRequest struct {
	Username  string    `json:"username" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	KeyBundle KeyBundle `json:"keyBundle" validate:"required"`
}

type SignUpResponse struct {
	UserID    string `json:"userId"`
	AuthToken string `json:"authToken"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignInResponse struct {
	UserID    string `json:"userId"`
	AuthToken string `json:"authToken"`
}
