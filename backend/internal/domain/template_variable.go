package domain

import (
	"time"

	"github.com/google/uuid"
)

type TemplateVariable struct {
	ID              uuid.UUID `json:"id"`
	TemplateID      uuid.UUID `json:"template_id"`
	Key             string    `json:"key"`
	Description     string    `json:"description"`
	VarType         string    `json:"var_type"`
	DefaultValue    string    `json:"default_value"`
	IsSensitive     bool      `json:"is_sensitive"`
	IsRequired      bool      `json:"is_required"`
	ValidationRegex string    `json:"validation_regex"`
	IsAutoParsed    bool      `json:"is_auto_parsed"`
	DisplayOrder    int       `json:"display_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type NewTemplateVariableParams struct {
	TemplateID      uuid.UUID
	Key             string
	Description     string
	VarType         string
	DefaultValue    string
	IsSensitive     bool
	IsRequired      bool
	ValidationRegex string
	IsAutoParsed    bool
}

func NewTemplateVariable(params NewTemplateVariableParams) *TemplateVariable {
	now := time.Now()
	varType := params.VarType
	if varType == "" {
		varType = "string"
	}
	return &TemplateVariable{
		ID:              uuid.New(),
		TemplateID:      params.TemplateID,
		Key:             params.Key,
		Description:     params.Description,
		VarType:         varType,
		DefaultValue:    params.DefaultValue,
		IsSensitive:     params.IsSensitive,
		IsRequired:      params.IsRequired,
		ValidationRegex: params.ValidationRegex,
		IsAutoParsed:    params.IsAutoParsed,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
