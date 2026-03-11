package contracts

import "github.com/google/uuid"

type (
	CreateTemplateVariable struct {
		TemplateID      uuid.UUID `json:"template_id" validate:"required,uuid4"`
		Key             string    `json:"key" validate:"required,min=1,max=255"`
		Description     string    `json:"description" validate:"max=1000"`
		VarType         string    `json:"var_type" validate:"omitempty,max=100"`
		DefaultValue    string    `json:"default_value"`
		IsSensitive     bool      `json:"is_sensitive"`
		IsRequired      bool      `json:"is_required"`
		ValidationRegex string    `json:"validation_regex" validate:"max=500"`
	}

	GetTemplateVariables struct {
		TemplateID uuid.UUID `json:"template_id" validate:"required,uuid4"`
	}

	UpdateTemplateVariable struct {
		ID              uuid.UUID `json:"id" validate:"required,uuid4"`
		Description     *string   `json:"description" validate:"omitempty,max=1000"`
		VarType         *string   `json:"var_type" validate:"omitempty,max=100"`
		DefaultValue    *string   `json:"default_value"`
		IsSensitive     *bool     `json:"is_sensitive"`
		IsRequired      *bool     `json:"is_required"`
		ValidationRegex *string   `json:"validation_regex" validate:"omitempty,max=500"`
		DisplayOrder    *int      `json:"display_order"`
	}

	DeleteTemplateVariable struct {
		ID uuid.UUID `json:"id" validate:"required,uuid4"`
	}

	ParseTemplateVariables struct {
		TemplateID uuid.UUID `json:"template_id" validate:"required,uuid4"`
	}
)
