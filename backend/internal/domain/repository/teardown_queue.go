package repository

import (
	"context"
	"time"

	"backend/internal/domain"
	"backend/pkg/errors"

	"github.com/google/uuid"
)

type TeardownQueueRepository interface {
	Enqueue(ctx context.Context, entry *domain.TeardownEntry) *errors.Error
	FindDue(ctx context.Context, now time.Time) (*domain.TeardownEntry, *errors.Error)
	UpdateStatus(ctx context.Context, envID uuid.UUID, status domain.TeardownStatus) *errors.Error
	ResetProcessing(ctx context.Context) *errors.Error
}
