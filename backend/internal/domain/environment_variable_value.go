package domain

import (
	"time"

	"github.com/google/uuid"
)

type EnvironmentVariableValue struct {
	ID                 uuid.UUID `json:"id"`
	EnvironmentID      uuid.UUID `json:"environment_id"`
	TemplateVariableID uuid.UUID `json:"template_variable_id"`
	Value              string    `json:"value"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func NewEnvironmentVariableValue(environmentID, templateVariableID uuid.UUID, value string) *EnvironmentVariableValue {
	now := time.Now()
	return &EnvironmentVariableValue{
		ID:                 uuid.New(),
		EnvironmentID:      environmentID,
		TemplateVariableID: templateVariableID,
		Value:              value,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
