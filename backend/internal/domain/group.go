package domain

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	WorkspaceID        uuid.UUID `json:"workspace_id"`
	AccessAllTemplates bool      `json:"access_all_templates"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func NewGroup(name, description string, workspaceID uuid.UUID, accessAllTemplates bool) *Group {
	now := time.Now()
	return &Group{
		ID:                 uuid.New(),
		Name:               name,
		Description:        description,
		WorkspaceID:        workspaceID,
		AccessAllTemplates: accessAllTemplates,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
