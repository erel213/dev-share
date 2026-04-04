package contracts

import "github.com/google/uuid"

type (
	CreateGroup struct {
		Name               string `json:"name" validate:"required,min=3,max=255"`
		Description        string `json:"description" validate:"omitempty,max=500"`
		AccessAllTemplates bool   `json:"access_all_templates"`
	}

	UpdateGroup struct {
		ID                 uuid.UUID `json:"id" validate:"required,uuid4"`
		Name               string    `json:"name" validate:"omitempty,min=3,max=255"`
		Description        *string   `json:"description" validate:"omitempty,max=500"`
		AccessAllTemplates *bool     `json:"access_all_templates"`
	}

	AddGroupMembers struct {
		UserIDs []uuid.UUID `json:"user_ids" validate:"required,dive,uuid4"`
	}

	AddGroupTemplateAccess struct {
		TemplateIDs []uuid.UUID `json:"template_ids" validate:"required,dive,uuid4"`
	}
)
