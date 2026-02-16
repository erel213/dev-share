package errors

import (
	"github.com/gofiber/fiber/v2"

	pkgerrors "backend/pkg/errors"
)

// Return wraps an error for handler return
// The error will be processed by the error handler middleware
func Return(err error) error {
	return err
}

// ReturnNotFound is a shorthand for common not found errors
func ReturnNotFound(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeNotFound, message).
		WithHTTPStatus(fiber.StatusNotFound).
		WithSeverity(pkgerrors.SeverityWarning)
}

// ReturnBadRequest is a shorthand for validation errors
func ReturnBadRequest(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeInvalidInput, message).
		WithHTTPStatus(fiber.StatusBadRequest).
		WithSeverity(pkgerrors.SeverityWarning)
}

// ReturnConflict is a shorthand for conflict errors
func ReturnConflict(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeConflict, message).
		WithHTTPStatus(fiber.StatusConflict).
		WithSeverity(pkgerrors.SeverityWarning)
}

// ReturnUnauthorized is a shorthand for unauthorized errors
func ReturnUnauthorized(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeUnauthorized, message).
		WithHTTPStatus(fiber.StatusUnauthorized).
		WithSeverity(pkgerrors.SeverityWarning)
}

// ReturnForbidden is a shorthand for forbidden errors
func ReturnForbidden(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeForbidden, message).
		WithHTTPStatus(fiber.StatusForbidden).
		WithSeverity(pkgerrors.SeverityWarning)
}

// ReturnInternalError is a shorthand for internal server errors
func ReturnInternalError(message string) error {
	return pkgerrors.WithCode(pkgerrors.CodeInternal, message).
		WithHTTPStatus(fiber.StatusInternalServerError).
		WithSeverity(pkgerrors.SeverityError)
}
