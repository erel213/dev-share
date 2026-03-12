package contracts

import "github.com/google/uuid"

type VariableValueEntry struct {
	TemplateVariableID uuid.UUID `json:"template_variable_id" validate:"required,uuid4"`
	Value              string    `json:"value" validate:"required"`
}

type (
	SetEnvironmentVariableValues struct {
		EnvironmentID uuid.UUID            `json:"environment_id" validate:"required,uuid4"`
		Values        []VariableValueEntry `json:"values" validate:"required,dive"`
	}

	GetEnvironmentVariableValues struct {
		EnvironmentID uuid.UUID `json:"environment_id" validate:"required,uuid4"`
	}
)
