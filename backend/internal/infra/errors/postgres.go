package errors

import (
	"database/sql"
	"fmt"
	"net/http"

	pkgerrors "backend/pkg/errors"
	"github.com/lib/pq"
)

// PostgreSQL error codes
const (
	CodePostgresUniqueViolation     = "23505"
	CodePostgresForeignKeyViolation = "23503"
	CodePostgresCheckViolation      = "23514"
	CodePostgresNotNullViolation    = "23502"
)

// WrapDatabaseError wraps database errors with appropriate categorization
// operation should describe the database operation being performed (e.g., "create_user", "get_workspace")
func WrapDatabaseError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Handle sql.ErrNoRows - this is a common expected error
	if err == sql.ErrNoRows {
		return pkgerrors.WithCode(pkgerrors.CodeNotFound, "record not found").
			WithMetadata("operation", operation).
			WithSeverity(pkgerrors.SeverityWarning) // Expected condition
	}

	// Handle PostgreSQL-specific errors
	if pqErr, ok := err.(*pq.Error); ok {
		return wrapPostgresError(pqErr, operation)
	}

	// Generic database error (unknown cause - needs investigation)
	// This captures stack trace since severity is Error
	return pkgerrors.Wrap(err, "database operation failed").
		WithMetadata("operation", operation).
		WithHTTPStatus(http.StatusInternalServerError).
		WithSeverity(pkgerrors.SeverityError)
}

// wrapPostgresError wraps PostgreSQL-specific errors with detailed context
func wrapPostgresError(pqErr *pq.Error, operation string) *pkgerrors.Error {
	base := pkgerrors.Wrap(pqErr, fmt.Sprintf("postgres error: %s", pqErr.Message)).
		WithMetadata("operation", operation).
		WithMetadata("pg_code", string(pqErr.Code)).
		WithMetadata("pg_detail", pqErr.Detail).
		WithMetadata("pg_hint", pqErr.Hint).
		WithMetadata("pg_constraint", pqErr.Constraint).
		WithMetadata("pg_table", pqErr.Table).
		WithMetadata("pg_column", pqErr.Column)

	switch pqErr.Code {
	case CodePostgresUniqueViolation:
		// Unique constraint violation - typically email, username, etc.
		return base.
			WithCode(pkgerrors.CodeConflict).
			WithHTTPStatus(http.StatusConflict).
			WithSeverity(pkgerrors.SeverityWarning) // Expected business condition

	case CodePostgresForeignKeyViolation:
		// Foreign key constraint violation - invalid reference
		return base.
			WithCode(pkgerrors.CodeInvalidInput).
			WithHTTPStatus(http.StatusBadRequest).
			WithSeverity(pkgerrors.SeverityWarning) // Expected validation failure

	case CodePostgresCheckViolation:
		// Check constraint violation - business rule violation
		return base.
			WithCode(pkgerrors.CodeValidation).
			WithHTTPStatus(http.StatusBadRequest).
			WithSeverity(pkgerrors.SeverityWarning)

	case CodePostgresNotNullViolation:
		// NOT NULL constraint violation - missing required field
		return base.
			WithCode(pkgerrors.CodeInvalidInput).
			WithHTTPStatus(http.StatusBadRequest).
			WithSeverity(pkgerrors.SeverityWarning)

	default:
		// Unknown PostgreSQL error - needs investigation, capture stack
		return base.
			WithCode(pkgerrors.CodeDatabase).
			WithHTTPStatus(http.StatusInternalServerError).
			WithSeverity(pkgerrors.SeverityError)
	}
}

// WrapTransactionError wraps transaction-related errors
func WrapTransactionError(err error, operation string) error {
	if err == nil {
		return nil
	}

	return pkgerrors.Wrap(err, "transaction failed").
		WithMetadata("operation", operation).
		WithHTTPStatus(http.StatusInternalServerError).
		WithSeverity(pkgerrors.SeverityError) // Transaction failures are serious
}

// WrapConnectionError wraps database connection errors
func WrapConnectionError(err error) error {
	if err == nil {
		return nil
	}

	return pkgerrors.Wrap(err, "database connection failed").
		WithCode(pkgerrors.CodeDatabase).
		WithHTTPStatus(http.StatusServiceUnavailable).
		WithSeverity(pkgerrors.SeverityCritical) // Connection failures are critical
}
