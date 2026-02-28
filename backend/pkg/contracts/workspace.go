package contracts

import "github.com/google/uuid"

type (
	CreateWorkspace struct {
		Name        string    `json:"name" validate:"required,min=3,max=100"`
		Description string    `json:"description" validate:"max=500"`
		AdminID     uuid.UUID `json:"admin_id" validate:"required,uuid4"`
	}

	UpdateWorkspace struct {
		ID          uuid.UUID `json:"id" validate:"required,uuid4"`
		Name        string    `json:"name" validate:"omitempty,min=3,max=100"`
		Description string    `json:"description" validate:"max=500"`
	}

	GetWorkspace struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	GetWorkspacesByAdmin struct {
		AdminID uuid.UUID `json:"admin_id" validate:"required,uuid4"`
	}

	ListWorkspaces struct {
		Limit  int    `json:"limit" validate:"omitempty,min=1,max=100"`
		Offset int    `json:"offset" validate:"omitempty,min=0"`
		SortBy string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=name created_at updated_at"`
		Order  string `json:"order" validate:"omitempty,oneof=ASC DESC"`
	}

	DeleteWorkspace struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}
)
