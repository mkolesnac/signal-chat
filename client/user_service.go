package main

import (
	"fmt"
	"net/http"
	"signal-chat/client/apiclient"
	"signal-chat/internal/api"
)

type UserService struct {
	apiClient apiclient.Client
}

func (u *UserService) GetUser(id string) (User, error) {
	requireNonEmpty("id", id)

	var resp api.GetUserResponse
	status, err := u.apiClient.Get(api.EndpointUser+"/"+id, &resp)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	if status != http.StatusOK {
		return User{}, fmt.Errorf("server returned error: %s", resp.Error)
	}

	return User{
		ID:       id,
		Username: resp.Username,
	}, nil
}
