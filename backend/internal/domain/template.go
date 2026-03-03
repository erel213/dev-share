package domain

import (
	"path/filepath"
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

func NewTemplate(name string, workspaceID uuid.UUID) *Template {
	now := time.Now()
	id := uuid.New()
	return &Template{
		ID:          id,
		Name:        name,
		WorkspaceID: workspaceID,
		Path:        filepath.Join(workspaceID.String(), id.String()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
