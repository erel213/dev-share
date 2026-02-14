package errors

import "net/http"

// Code represents an error code for categorizing errors
type Code string

const (
	// CodeUnknown represents an unknown or unclassified error
	CodeUnknown Code = "UNKNOWN"
	// CodeInternal represents an internal system error
	CodeInternal Code = "INTERNAL"
	// CodeInvalidInput represents invalid input or validation failure
	CodeInvalidInput Code = "INVALID_INPUT"
	// CodeNotFound represents a resource that was not found
	CodeNotFound Code = "NOT_FOUND"
	// CodeConflict represents a conflict with existing data (e.g., unique constraint)
	CodeConflict Code = "CONFLICT"
	// CodeUnauthorized represents an authentication failure
	CodeUnauthorized Code = "UNAUTHORIZED"
	// CodeForbidden represents insufficient permissions
	CodeForbidden Code = "FORBIDDEN"
	// CodeDatabase represents a generic database error
	CodeDatabase Code = "DATABASE_ERROR"
	// CodeConstraint represents a database constraint violation
	CodeConstraint Code = "CONSTRAINT_VIOLATION"
	// CodeValidation represents a validation error
	CodeValidation Code = "VALIDATION_ERROR"
)

// HTTPStatus returns the HTTP status code for this error code
func (c Code) HTTPStatus() int {
	switch c {
	case CodeInvalidInput, CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeConstraint:
		return http.StatusConflict
	case CodeInternal, CodeDatabase, CodeUnknown:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// String returns the string representation of the error code
func (c Code) String() string {
	return string(c)
}
