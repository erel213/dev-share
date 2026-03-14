package domain

import (
	"path/filepath"
	"time"

	pkgerrors "backend/pkg/errors"

	"github.com/google/uuid"
)

type Template struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name" validate:"required,min=3,max=255"`
	WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
	Path        string    `json:"path" validate:"required,filepath"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewTemplate(name string, workspaceID uuid.UUID, validator Validator) (*Template, *pkgerrors.Error) {
	now := time.Now()
	id := uuid.New()
	t := &Template{
		ID:          id,
		Name:        name,
		WorkspaceID: workspaceID,
		Path:        filepath.Join(workspaceID.String(), id.String()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := validator.Validate(t); err != nil {
		return nil, err
	}

	return t, nil
}
