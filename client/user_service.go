package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"signal-chat/client/models"
	"signal-chat/internal/api"
)

type UserAPI interface {
	Get(route string) (int, []byte, error)
}

type UserService struct {
	apiClient UserAPI
}

func NewUserService(apiClient UserAPI) *UserService {
	return &UserService{apiClient: apiClient}
}

func (u *UserService) GetUser(id string) (models.User, error) {
	panicIfEmpty("id", id)

	status, body, err := u.apiClient.Get(api.EndpointUser + "/" + id)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	if status != http.StatusOK {
		return models.User{}, fmt.Errorf("server returned unsuccessful status code: %v", status)
	}
	var resp api.GetUserResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return models.User{
		ID:       id,
		Username: resp.Username,
	}, nil
}
