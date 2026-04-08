package contracts

import (
	"time"

	"github.com/google/uuid"
)

type (
	CreateEnvironment struct {
		Name        string    `json:"name" validate:"required,min=3,max=255"`
		Description string    `json:"description" validate:"omitempty,max=1000"`
		TemplateID  uuid.UUID `json:"template_id" validate:"required,uuid4"`
		TTLSeconds  *int      `json:"ttl_seconds" validate:"omitempty,min=60"`
	}

	GetEnvironment struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	ListEnvironments struct {
		Scope      string `query:"scope" validate:"omitempty,oneof=user all"`
		Status     string `query:"status" validate:"omitempty"`
		TemplateID string `query:"template_id" validate:"omitempty"`
		CreatedBy  string `query:"created_by" validate:"omitempty"`
		Search     string `query:"search" validate:"omitempty,max=255"`
		SortBy     string `query:"sort_by" validate:"omitempty,oneof=created_at updated_at name status"`
		Order      string `query:"order" validate:"omitempty,oneof=ASC DESC"`
		Limit      int    `query:"limit" validate:"omitempty,min=1,max=100"`
		Offset     int    `query:"offset" validate:"omitempty,min=0"`
	}

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

	EnvironmentResponse struct {
		ID            uuid.UUID  `json:"id"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
		CreatedBy     uuid.UUID  `json:"created_by"`
		CreatedByName string     `json:"created_by_name"`
		WorkspaceID   uuid.UUID  `json:"workspace_id"`
		TemplateID    uuid.UUID  `json:"template_id"`
		TemplateName  string     `json:"template_name"`
		Status        string     `json:"status"`
		LastAppliedAt *time.Time `json:"last_applied_at,omitempty"`
		LastOperation string     `json:"last_operation,omitempty"`
		LastError     string     `json:"last_error,omitempty"`
		TTLSeconds    *int       `json:"ttl_seconds,omitempty"`
		CreatedAt     time.Time  `json:"created_at"`
		UpdatedAt     time.Time  `json:"updated_at"`
	}
)
