package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewTemplateVariableDefaults(t *testing.T) {
	templateID := uuid.New()
	v := NewTemplateVariable(NewTemplateVariableParams{
		TemplateID: templateID,
		Key:        "test_key",
		IsRequired: true,
	})

	if v.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if v.TemplateID != templateID {
		t.Errorf("expected template ID %s, got %s", templateID, v.TemplateID)
	}
	if v.Key != "test_key" {
		t.Errorf("expected key 'test_key', got %q", v.Key)
	}
	if v.VarType != "string" {
		t.Errorf("expected default var_type 'string', got %q", v.VarType)
	}
	if !v.IsRequired {
		t.Error("expected IsRequired=true")
	}
	if v.IsSensitive {
		t.Error("expected IsSensitive=false by default")
	}
	if v.IsAutoParsed {
		t.Error("expected IsAutoParsed=false")
	}
	if v.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if v.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNewTemplateVariableWithAllParams(t *testing.T) {
	templateID := uuid.New()
	v := NewTemplateVariable(NewTemplateVariableParams{
		TemplateID:      templateID,
		Key:             "db_password",
		Description:     "Database password",
		VarType:         "string",
		DefaultValue:    "default123",
		IsSensitive:     true,
		IsRequired:      false,
		ValidationRegex: `^[a-zA-Z0-9]+$`,
		IsAutoParsed:    true,
	})

	if v.Description != "Database password" {
		t.Errorf("expected description 'Database password', got %q", v.Description)
	}
	if v.DefaultValue != "default123" {
		t.Errorf("expected default_value 'default123', got %q", v.DefaultValue)
	}
	if !v.IsSensitive {
		t.Error("expected IsSensitive=true")
	}
	if v.IsRequired {
		t.Error("expected IsRequired=false")
	}
	if v.ValidationRegex != `^[a-zA-Z0-9]+$` {
		t.Errorf("expected validation regex, got %q", v.ValidationRegex)
	}
	if !v.IsAutoParsed {
		t.Error("expected IsAutoParsed=true")
	}
}

func TestNewTemplateVariableEmptyVarTypeDefaultsToString(t *testing.T) {
	v := NewTemplateVariable(NewTemplateVariableParams{
		TemplateID: uuid.New(),
		Key:        "test",
		VarType:    "",
	})
	if v.VarType != "string" {
		t.Errorf("expected 'string' for empty VarType, got %q", v.VarType)
	}
}
