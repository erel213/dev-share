package domain

import (
	"time"

	"github.com/google/uuid"
)

type TeardownStatus string

const (
	TeardownStatusPending    TeardownStatus = "pending"
	TeardownStatusProcessing TeardownStatus = "processing"
	TeardownStatusCompleted  TeardownStatus = "completed"
)

type TeardownEntry struct {
	EnvironmentID uuid.UUID
	TeardownAt    time.Time
	Status        TeardownStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
