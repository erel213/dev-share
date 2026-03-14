package domain

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type EnvironmentStatus string

const (
	EnvironmentStatusPending     EnvironmentStatus = "pending"
	EnvironmentStatusInitialized EnvironmentStatus = "initialized"
	EnvironmentStatusPlanning    EnvironmentStatus = "planning"
	EnvironmentStatusApplying    EnvironmentStatus = "applying"
	EnvironmentStatusReady       EnvironmentStatus = "ready"
	EnvironmentStatusDestroying  EnvironmentStatus = "destroying"
	EnvironmentStatusDestroyed   EnvironmentStatus = "destroyed"
	EnvironmentStatusError       EnvironmentStatus = "error"
)

// OperationBlockingStatuses are statuses that indicate an operation is in progress
// and concurrent operations must be rejected.
var OperationBlockingStatuses = []EnvironmentStatus{
	EnvironmentStatusApplying,
	EnvironmentStatusDestroying,
	EnvironmentStatusPlanning,
}

func (s EnvironmentStatus) IsValid() bool {
	switch s {
	case EnvironmentStatusPending, EnvironmentStatusInitialized, EnvironmentStatusPlanning,
		EnvironmentStatusApplying, EnvironmentStatusReady, EnvironmentStatusDestroying,
		EnvironmentStatusDestroyed, EnvironmentStatusError:
		return true
	}
	return false
}

func (s EnvironmentStatus) String() string { return string(s) }

type Environment struct {
	ID            uuid.UUID         `json:"id"`
	Name          string            `json:"name"`
	CreatedAt     time.Time         `json:"created_at"`
	CreatedBy     uuid.UUID         `json:"created_by"`
	Description   string            `json:"description"`
	WorkspaceID   uuid.UUID         `json:"workspace_id"`
	TemplateID    uuid.UUID         `json:"template_id"`
	Status        EnvironmentStatus `json:"status"`
	LastAppliedAt *time.Time        `json:"last_applied_at,omitempty"`
	LastOperation string            `json:"last_operation,omitempty"`
	LastError     string            `json:"last_error,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func NewEnvironment(name, description string, createdBy, workspaceID, templateId uuid.UUID) *Environment {
	return &Environment{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		WorkspaceID: workspaceID,
		Status:      EnvironmentStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		TemplateID:  templateId,
	}
}

func (e Environment) ExecutionPath() string {
	return filepath.Join(e.WorkspaceID.String(), e.ID.String())
}

// CanStartOperation checks whether the environment can accept a new Terraform
// operation. Returns an error describing why not if the status is blocking.
func (e *Environment) CanStartOperation() error {
	for _, s := range OperationBlockingStatuses {
		if e.Status == s {
			return fmt.Errorf("environment is currently %s — operation in progress", e.Status)
		}
	}
	if e.Status == EnvironmentStatusDestroyed {
		return fmt.Errorf("environment has been destroyed")
	}
	return nil
}

func OperationFromStatus(s EnvironmentStatus) string {
	switch s {
	case EnvironmentStatusPlanning:
		return "plan"
	case EnvironmentStatusApplying:
		return "apply"
	case EnvironmentStatusDestroying:
		return "destroy"
	default:
		return string(s)
	}
}

func OperationFromStatus(s EnvironmentStatus) string {
	switch s {
	case EnvironmentStatusPlanning:
		return "plan"
	case EnvironmentStatusApplying:
		return "apply"
	case EnvironmentStatusDestroying:
		return "destroy"
	default:
		return string(s)
	}
}
