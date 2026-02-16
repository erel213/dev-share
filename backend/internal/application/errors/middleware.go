package errors

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	pkgerrors "backend/pkg/errors"
)

// ErrorHandler returns a Fiber error handler that converts errors to JSON responses
// This should be configured in fiber.Config.ErrorHandler
func ErrorHandler() func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		if err == nil {
			return nil
		}

		// Convert to application error
		var appErr *pkgerrors.Error
		if !errors.As(err, &appErr) {
			// Unknown error - wrap it with stack trace
			appErr = pkgerrors.Wrap(err, "internal server error").
				WithHTTPStatus(fiber.StatusInternalServerError).
				WithSeverity(pkgerrors.SeverityError)
		}

		// Log error with structured context
		logError(c, appErr)

		// Return JSON error response
		return c.Status(appErr.HTTPStatus()).JSON(ErrorResponse{
			Error: ErrorDetail{
				Code:     string(appErr.Code()),
				Message:  appErr.Error(),
				Metadata: appErr.GetMetadata(),
			},
		})
	}
}

// logError logs the error with structured context
func logError(c *fiber.Ctx, err *pkgerrors.Error) {
	// Build log attributes
	attrs := []any{
		"error", err, // Uses LogValue() for structured output
		"path", c.Path(),
		"method", c.Method(),
		"status", err.HTTPStatus(),
		"severity", err.Severity(),
		"code", err.Code(),
	}

	// Add request ID if available
	if reqID := c.Get("X-Request-ID"); reqID != "" {
		attrs = append(attrs, "request_id", reqID)
	}

	// Log at appropriate level based on severity
	switch err.Severity() {
	case pkgerrors.SeverityDebug:
		slog.Debug("request error", attrs...)
	case pkgerrors.SeverityInfo:
		slog.Info("request error", attrs...)
	case pkgerrors.SeverityWarning:
		slog.Warn("request error", attrs...)
	case pkgerrors.SeverityError:
		slog.Error("request error", attrs...)
	case pkgerrors.SeverityCritical:
		slog.Error("CRITICAL request error", attrs...)
	default:
		slog.Error("request error", attrs...)
	}
}

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the error information returned to clients
type ErrorDetail struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
