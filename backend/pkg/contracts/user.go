package contracts

import "github.com/google/uuid"

type (
	CreateLocalUser struct {
		Name        string    `json:"name" validate:"required"`
		Email       string    `json:"email" validate:"required,email"`
		Password    string    `json:"password" validate:"required,min=8"`
		WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
	}
)
