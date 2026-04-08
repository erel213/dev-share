package repository

import (
	"backend/internal/domain/errors"
	pkgerrors "backend/pkg/errors"

	"github.com/google/uuid"
)

type ListOptions struct {
	Limit    int
	Offset   int
	SortBy   string
	Order    string // "ASC" or "DESC"
	FilterBy map[string]string
}

func (o *ListOptions) Validate() *pkgerrors.Error {
	if o.Limit < 0 {
		return errors.ErrInvalidInput
	}
	if o.Offset < 0 {
		return errors.ErrInvalidInput
	}
	if o.Order != "" && o.Order != "ASC" && o.Order != "DESC" {
		return errors.ErrInvalidInput
	}
	return nil
}

func (o *ListOptions) ApplyDefaults() {
	if o.Limit == 0 {
		o.Limit = 50 // Default page size
	}
	if o.Order == "" {
		o.Order = "DESC"
	}
	if o.SortBy == "" {
		o.SortBy = "created_at"
	}
}

// EnvironmentListOptions holds filters for the enriched environment listing.
type EnvironmentListOptions struct {
	WorkspaceID uuid.UUID
	CreatorIDs  []uuid.UUID // scope=user: self + co-members; scope=all: empty (no filter)
	Statuses    []string
	TemplateID  *uuid.UUID
	Search      string
	SortBy      string
	Order       string
	Limit       int
	Offset      int
}
