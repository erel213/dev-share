package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OauthProviderGitHub OauthProvider = "github"
)

type (
	User struct {
		ID            uuid.UUID     `json:"id"`
		OauthProvider OauthProvider `json:"oauth_provider"`
		OauthID       string        `json:"oauth_id"`
		Name          string        `json:"name"`
		Email         string        `json:"email"`
		WorkspaceID   uuid.UUID     `json:"workspace_id"`
		CreatedAt     time.Time     `json:"created_at"`
		UpdatedAt     time.Time     `json:"updated_at"`
	}

	OauthProvider string
)

func NewUser(oauthProvider OauthProvider, oauthID, name, email string, workspaceId uuid.UUID) *User {
	return &User{
		OauthProvider: oauthProvider,
		OauthID:       oauthID,
		Name:          name,
		Email:         email,
		WorkspaceID:   workspaceId,
	}
}
