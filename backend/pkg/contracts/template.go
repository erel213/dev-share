package contracts

import "github.com/google/uuid"

type (
	CreateTemplate struct {
		Name        string    `form:"name" validate:"required,min=3,max=255"`
		WorkspaceID uuid.UUID `form:"workspace_id" validate:"required,uuid4"`
	}

	UpdateTemplate struct {
		ID   uuid.UUID `form:"id" validate:"required,uuid4"`
		Name string    `form:"name" validate:"omitempty,min=3,max=255"`
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
