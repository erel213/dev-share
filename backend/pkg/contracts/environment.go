package contracts

import "github.com/google/uuid"

type (
	CreateEnvironment struct {
		Name        string    `json:"name" validate:"required,min=3,max=255"`
		Description string    `json:"description" validate:"omitempty,max=1000"`
		TemplateID  uuid.UUID `json:"template_id" validate:"required,uuid4"`
	}

	GetEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	GetEnvironmentsByWorkspace struct{}

	ApplyEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	PlanEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	DestroyEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	DeleteEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}
)
