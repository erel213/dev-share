package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewEnvironment(t *testing.T) {
	name := "my-env"
	description := "test environment"
	createdBy := uuid.New()
	workspaceID := uuid.New()
	templateID := uuid.New()

	env := NewEnvironment(name, description, createdBy, workspaceID, templateID)

	if env.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if env.Name != name {
		t.Errorf("expected name %q, got %q", name, env.Name)
	}
	if env.Description != description {
		t.Errorf("expected description %q, got %q", description, env.Description)
	}
	if env.CreatedBy != createdBy {
		t.Errorf("expected created_by %v, got %v", createdBy, env.CreatedBy)
	}
	if env.WorkspaceID != workspaceID {
		t.Errorf("expected workspace_id %v, got %v", workspaceID, env.WorkspaceID)
	}
	if env.TemplateID != templateID {
		t.Errorf("expected template_id %v, got %v", templateID, env.TemplateID)
	}
	if env.Status != EnvironmentStatusPending {
		t.Errorf("expected status %q, got %q", EnvironmentStatusPending, env.Status)
	}
}

func TestEnvironmentStatus_IsValid(t *testing.T) {
	valid := []EnvironmentStatus{
		EnvironmentStatusPending,
		EnvironmentStatusInitialized,
		EnvironmentStatusPlanning,
		EnvironmentStatusApplying,
		EnvironmentStatusReady,
		EnvironmentStatusDestroying,
		EnvironmentStatusDestroyed,
		EnvironmentStatusError,
	}

	for _, s := range valid {
		t.Run("valid_"+string(s), func(t *testing.T) {
			if !s.IsValid() {
				t.Errorf("expected %q to be valid", s)
			}
		})
	}

	invalid := []EnvironmentStatus{"unknown", "", "running", "PENDING"}
	for _, s := range invalid {
		t.Run("invalid_"+string(s), func(t *testing.T) {
			if s.IsValid() {
				t.Errorf("expected %q to be invalid", s)
			}
		})
	}
}

func TestEnvironmentStatus_String(t *testing.T) {
	s := EnvironmentStatusApplying
	if s.String() != "applying" {
		t.Errorf("expected %q, got %q", "applying", s.String())
	}
}

func TestEnvironment_ExecutionPath(t *testing.T) {
	workspaceID := uuid.New()
	envID := uuid.New()
	env := &Environment{
		ID:          envID,
		WorkspaceID: workspaceID,
	}

	expected := workspaceID.String() + "/" + envID.String()
	if env.ExecutionPath() != expected {
		t.Errorf("expected %q, got %q", expected, env.ExecutionPath())
	}
}

func TestEnvironment_CanStartOperation(t *testing.T) {
	tests := []struct {
		name      string
		status    EnvironmentStatus
		expectErr bool
	}{
		{"pending allows operation", EnvironmentStatusPending, false},
		{"initialized allows operation", EnvironmentStatusInitialized, false},
		{"ready allows operation", EnvironmentStatusReady, false},
		{"error allows retry", EnvironmentStatusError, false},
		{"applying blocks operation", EnvironmentStatusApplying, true},
		{"destroying blocks operation", EnvironmentStatusDestroying, true},
		{"planning blocks operation", EnvironmentStatusPlanning, true},
		{"destroyed blocks operation", EnvironmentStatusDestroyed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{Status: tt.status}
			err := env.CanStartOperation()

			if tt.expectErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestOperationBlockingStatuses(t *testing.T) {
	expected := map[EnvironmentStatus]bool{
		EnvironmentStatusApplying:   true,
		EnvironmentStatusDestroying: true,
		EnvironmentStatusPlanning:   true,
	}

	if len(OperationBlockingStatuses) != len(expected) {
		t.Fatalf("expected %d blocking statuses, got %d", len(expected), len(OperationBlockingStatuses))
	}

	for _, s := range OperationBlockingStatuses {
		if !expected[s] {
			t.Errorf("unexpected blocking status: %q", s)
		}
	}
}
