package data

import (
	"encoding/json"
	"fmt"

	"github.com/dlvhdr/gh-dash/v4/internal/providers"
)

type gitlabCurrentUser struct {
	Username string `json:"username"`
}

func GitLabCurrentUser(provider providers.Instance) (string, error) {
	body, _, err := gitlabGet(provider, "/user", map[string]string{})
	if err != nil {
		return "", err
	}
	var user gitlabCurrentUser
	if err := json.Unmarshal(body, &user); err != nil {
		return "", err
	}
	if user.Username == "" {
		return "", fmt.Errorf("gitlab current user response missing username")
	}
	return user.Username, nil
}
