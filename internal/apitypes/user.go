package apitypes

type GetUserResponse struct {
	User User `json:"user"`
}

type GetAllUsersResponse struct {
	Users []User `json:"users"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
