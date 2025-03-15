package api

type GetUserResponse struct {
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

type GetPrekeyBundleResponse struct {
	PreKeyBundle PreKeyBundle `json:"preKeyBundle"`
	Error        string       `json:"error,omitempty"`
}
