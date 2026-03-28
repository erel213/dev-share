package contracts

import "github.com/google/uuid"

type (
	CreateLocalUser struct {
		Name        string    `json:"name" validate:"required,min=2,max=100"`
		Email       string    `json:"email" validate:"required,email"`
		Password    string    `json:"password" validate:"required,min=8,strongpassword"`
		WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
	}

	LoginLocalUser struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	LoginResponse struct {
		UserID      uuid.UUID `json:"user_id"`
		Name        string    `json:"name"`
		IsAdmin     bool      `json:"is_admin"`
		WorkspaceID uuid.UUID `json:"workspace_id"`
	}
)
