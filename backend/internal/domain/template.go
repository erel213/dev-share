package domain

import "github.com/google/uuid"

type Template struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Path        string    `json:"path"`
}

func NewTemplate(name string, workspaceID uuid.UUID, path string) *Template {
	return &Template{
		ID:          uuid.New(),
		Name:        name,
		WorkspaceID: workspaceID,
		Path:        path,
	}
}
