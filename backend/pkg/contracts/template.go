package contracts

import "github.com/google/uuid"

type (
	CreateTemplate struct {
		Name        string    `json:"name" validate:"required,min=3,max=255"`
		WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
		Path        string    `json:"path" validate:"required,min=1"`
	}

	UpdateTemplate struct {
		ID   uuid.UUID `json:"id" validate:"required,uuid4"`
		Name string    `json:"name" validate:"omitempty,min=3,max=255"`
		Path string    `json:"path" validate:"omitempty,min=1"`
	}

	GetTemplate struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	GetTemplatesByWorkspace struct {
		WorkspaceID uuid.UUID `json:"workspace_id" validate:"required,uuid4"`
	}

	ListTemplates struct {
		Limit  int    `json:"limit" validate:"omitempty,min=1,max=100"`
		Offset int    `json:"offset" validate:"omitempty,min=0"`
		SortBy string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=name created_at updated_at"`
		Order  string `json:"order" validate:"omitempty,oneof=ASC DESC"`
	}

	DeleteTemplate struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}
)
