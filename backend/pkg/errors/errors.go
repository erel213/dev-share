package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Error represents a rich error with metadata, severity, and optional stack trace
type Error struct {
	message    string
	code       Code
	severity   Severity
	cause      error
	metadata   map[string]interface{}
	timestamp  time.Time
	stack      []uintptr
	httpStatus int
}

// New creates a new error with the given message
func New(message string) *Error {
	return &Error{
		message:    message,
		code:       CodeUnknown,
		severity:   SeverityError,
		timestamp:  time.Now(),
		httpStatus: http.StatusInternalServerError,
		metadata:   make(map[string]interface{}),
		stack:      captureStack(3), // Skip: runtime.Callers, captureStack, New
	}
}

// Newf creates a new error with a formatted message
func Newf(format string, args ...interface{}) *Error {
	return &Error{
		message:    fmt.Sprintf(format, args...),
		code:       CodeUnknown,
		severity:   SeverityError,
		timestamp:  time.Now(),
		httpStatus: http.StatusInternalServerError,
		metadata:   make(map[string]interface{}),
		stack:      captureStack(3),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	// If already an Error, preserve its properties but add context
	if appErr, ok := err.(*Error); ok {
		return &Error{
			message:    message,
			code:       appErr.code,
			severity:   appErr.severity,
			cause:      appErr,
			metadata:   copyMetadata(appErr.metadata),
			timestamp:  time.Now(),
			stack:      appErr.stack, // Preserve original stack
			httpStatus: appErr.httpStatus,
		}
	}

	// Wrap a regular error
	return &Error{
		message:    message,
		code:       CodeUnknown,
		severity:   SeverityError,
		cause:      err,
		timestamp:  time.Now(),
		httpStatus: http.StatusInternalServerError,
		metadata:   make(map[string]interface{}),
		stack:      captureStack(3),
	}
}

// Wrapf wraps an existing error with a formatted message
func Wrapf(err error, format string, args ...interface{}) *Error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WithCode creates a new error with a specific error code
func WithCode(code Code, message string) *Error {
	severity := SeverityError
	// Adjust default severity based on code
	switch code {
	case CodeNotFound, CodeConflict, CodeInvalidInput, CodeValidation:
		severity = SeverityWarning
	case CodeUnauthorized, CodeForbidden:
		severity = SeverityWarning
	}

	var stack []uintptr
	if severity.ShouldCaptureStack() {
		stack = captureStack(3)
	}

	return &Error{
		message:    message,
		code:       code,
		severity:   severity,
		timestamp:  time.Now(),
		httpStatus: code.HTTPStatus(),
		metadata:   make(map[string]interface{}),
		stack:      stack,
	}
}

// WithCodef creates a new error with a specific error code and formatted message
func WithCodef(code Code, format string, args ...interface{}) *Error {
	return WithCode(code, fmt.Sprintf(format, args...))
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Unwrap returns the underlying cause for error unwrapping
func (e *Error) Unwrap() error {
	return e.cause
}

// Is implements error comparison for errors.Is()
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.code == t.code
}

// As implements type assertion for errors.As()
func (e *Error) As(target interface{}) bool {
	if t, ok := target.(**Error); ok {
		*t = e
		return true
	}
	return false
}

// WithMetadata adds metadata to the error
func (e *Error) WithMetadata(key string, value interface{}) *Error {
	if e.metadata == nil {
		e.metadata = make(map[string]interface{})
	}
	e.metadata[key] = value
	return e
}

// WithHTTPStatus sets the HTTP status code
func (e *Error) WithHTTPStatus(status int) *Error {
	e.httpStatus = status
	return e
}

// WithSeverity sets the severity level
func (e *Error) WithSeverity(severity Severity) *Error {
	e.severity = severity

	// Capture stack if severity changed to Error/Critical and not already captured
	if severity.ShouldCaptureStack() && len(e.stack) == 0 {
		e.stack = captureStack(3)
	}

	return e
}

// WithCode sets the error code
func (e *Error) WithCode(code Code) *Error {
	e.code = code
	// Update HTTP status to match the code if not already set
	if e.httpStatus == 0 || e.httpStatus == code.HTTPStatus() {
		e.httpStatus = code.HTTPStatus()
	}
	return e
}

// Code returns the error code
func (e *Error) Code() Code {
	return e.code
}

// Severity returns the severity level
func (e *Error) Severity() Severity {
	return e.severity
}

// HTTPStatus returns the HTTP status code
func (e *Error) HTTPStatus() int {
	return e.httpStatus
}

// GetMetadata returns a copy of the metadata
func (e *Error) GetMetadata() map[string]interface{} {
	return copyMetadata(e.metadata)
}

// StackTrace returns the formatted stack trace
func (e *Error) StackTrace() []Frame {
	return formatStack(e.stack)
}

// Timestamp returns the error creation timestamp
func (e *Error) Timestamp() time.Time {
	return e.timestamp
}

// LogValue implements slog.LogValuer for structured logging integration
func (e *Error) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("message", e.message),
		slog.String("code", string(e.code)),
		slog.String("severity", e.severity.String()),
		slog.Time("timestamp", e.timestamp),
	}

	if len(e.metadata) > 0 {
		metadataAttrs := make([]slog.Attr, 0, len(e.metadata))
		for k, v := range e.metadata {
			metadataAttrs = append(metadataAttrs, slog.Any(k, v))
		}
		attrs = append(attrs, slog.Any("metadata", slog.GroupValue(metadataAttrs...)))
	}

	if e.cause != nil {
		attrs = append(attrs, slog.String("cause", e.cause.Error()))
	}

	if len(e.stack) > 0 {
		frames := e.StackTrace()
		if len(frames) > 0 {
			stackAttrs := make([]slog.Attr, 0, len(frames))
			for i, frame := range frames {
				stackAttrs = append(stackAttrs, slog.Any(fmt.Sprintf("frame_%d", i), slog.GroupValue(
					slog.String("file", frame.File),
					slog.Int("line", frame.Line),
					slog.String("function", frame.Function),
				)))
			}
			attrs = append(attrs, slog.Any("stack_trace", slog.GroupValue(stackAttrs...)))
		}
	}

	return slog.GroupValue(attrs...)
}

// copyMetadata creates a copy of the metadata map
func copyMetadata(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	copy := make(map[string]interface{}, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.code == CodeNotFound
	}
	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.code == CodeConflict
	}
	return false
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.code == CodeInvalidInput || appErr.code == CodeValidation
	}
	return false
}
