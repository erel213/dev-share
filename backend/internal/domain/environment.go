package domain

import (
	"time"

	"github.com/google/uuid"
)

type Environment struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Description string    `json:"description"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	TemplateID  uuid.UUID `json:"template_id"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewEnvironment(name, description string, createdBy, workspaceID uuid.UUID) *Environment {
	return &Environment{
		Name:        name,
		CreatedBy:   createdBy,
		Description: description,
		WorkspaceID: workspaceID,
	}
}
