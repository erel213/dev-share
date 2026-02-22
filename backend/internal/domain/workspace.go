package domain

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AdminID     *uuid.UUID `json:"admin"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewWorkspace(name string, description string, adminId *uuid.UUID) *Workspace {
	return &Workspace{
		Name:        name,
		Description: description,
		AdminID:     adminId,
	}
}
