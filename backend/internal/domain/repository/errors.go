package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("entity not found")
	ErrConflict     = errors.New("entity already exists")
	ErrInvalidInput = errors.New("invalid input")
)

type NotFoundError struct {
	EntityType string
	ID         uuid.UUID
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.EntityType, e.ID)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

type ConflictError struct {
	EntityType string
	Field      string
	Value      string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s already exists with %s: %s", e.EntityType, e.Field, e.Value)
}

func (e *ConflictError) Is(target error) bool {
	return target == ErrConflict
}
