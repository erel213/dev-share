package errors

import (
	"fmt"
	"net/http"

	pkgerrors "backend/pkg/errors"
	"github.com/google/uuid"
)

// Sentinel errors for backward compatibility
var (
	// ErrNotFound is a generic not found error
	ErrNotFound = pkgerrors.WithCode(pkgerrors.CodeNotFound, "entity not found")
	// ErrConflict is a generic conflict error
	ErrConflict = pkgerrors.WithCode(pkgerrors.CodeConflict, "entity already exists")
	// ErrInvalidInput is a generic invalid input error
	ErrInvalidInput = pkgerrors.WithCode(pkgerrors.CodeInvalidInput, "invalid input")
)

// NotFound creates a domain NotFound error with entity context
func NotFound(entityType string, id uuid.UUID) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeNotFound,
		fmt.Sprintf("%s not found: %s", entityType, id),
	).
		WithMetadata("entity_type", entityType).
		WithMetadata("entity_id", id.String()).
		WithHTTPStatus(http.StatusNotFound).
		WithSeverity(pkgerrors.SeverityWarning) // Not found is expected, not critical
}

// NotFoundByField creates a domain NotFound error for non-UUID lookups
func NotFoundByField(entityType, field, value string) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeNotFound,
		fmt.Sprintf("%s not found with %s: %s", entityType, field, value),
	).
		WithMetadata("entity_type", entityType).
		WithMetadata("field", field).
		WithMetadata("value", value).
		WithHTTPStatus(http.StatusNotFound).
		WithSeverity(pkgerrors.SeverityWarning)
}

// Conflict creates a domain Conflict error
func Conflict(entityType, field, value string) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeConflict,
		fmt.Sprintf("%s already exists with %s: %s", entityType, field, value),
	).
		WithMetadata("entity_type", entityType).
		WithMetadata("field", field).
		WithMetadata("value", value).
		WithHTTPStatus(http.StatusConflict).
		WithSeverity(pkgerrors.SeverityWarning) // Conflicts are expected, not critical
}

// InvalidInput creates a validation error
func InvalidInput(field, reason string) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeInvalidInput,
		fmt.Sprintf("invalid input for %s: %s", field, reason),
	).
		WithMetadata("field", field).
		WithMetadata("reason", reason).
		WithHTTPStatus(http.StatusBadRequest).
		WithSeverity(pkgerrors.SeverityWarning) // Validation errors are expected
}

// ValidationError creates a general validation error with multiple field errors
func ValidationError(message string, fieldErrors map[string]string) *pkgerrors.Error {
	err := pkgerrors.WithCode(
		pkgerrors.CodeValidation,
		message,
	).
		WithHTTPStatus(http.StatusBadRequest).
		WithSeverity(pkgerrors.SeverityWarning)

	// Add each field error as metadata
	for field, reason := range fieldErrors {
		err = err.WithMetadata(fmt.Sprintf("field_%s", field), reason)
	}

	return err
}

// Unauthorized creates an unauthorized error
func Unauthorized(reason string) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeUnauthorized,
		fmt.Sprintf("unauthorized: %s", reason),
	).
		WithMetadata("reason", reason).
		WithHTTPStatus(http.StatusUnauthorized).
		WithSeverity(pkgerrors.SeverityWarning)
}

// Forbidden creates a forbidden error
func Forbidden(resource, action string) *pkgerrors.Error {
	return pkgerrors.WithCode(
		pkgerrors.CodeForbidden,
		fmt.Sprintf("forbidden: insufficient permissions to %s %s", action, resource),
	).
		WithMetadata("resource", resource).
		WithMetadata("action", action).
		WithHTTPStatus(http.StatusForbidden).
		WithSeverity(pkgerrors.SeverityWarning)
}
