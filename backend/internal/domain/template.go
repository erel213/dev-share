package domain

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTemplate(name string, workspaceID uuid.UUID, path string) *Template {
	now := time.Now()
	return &Template{
		ID:          uuid.New(),
		Name:        name,
		WorkspaceID: workspaceID,
		Path:        path,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
