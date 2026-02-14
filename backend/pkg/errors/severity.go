package errors

import "log/slog"

// Severity represents the severity level of an error
type Severity int

const (
	// SeverityDebug represents debug-level errors (lowest severity)
	SeverityDebug Severity = iota
	// SeverityInfo represents informational errors
	SeverityInfo
	// SeverityWarning represents warning-level errors (expected conditions)
	SeverityWarning
	// SeverityError represents error-level conditions (stack trace captured)
	SeverityError
	// SeverityCritical represents critical errors requiring immediate attention (stack trace captured)
	SeverityCritical
)

// String returns the string representation of the severity level
func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// LogLevel converts Severity to slog.Level for structured logging integration
func (s Severity) LogLevel() slog.Level {
	switch s {
	case SeverityDebug:
		return slog.LevelDebug
	case SeverityInfo:
		return slog.LevelInfo
	case SeverityWarning:
		return slog.LevelWarn
	case SeverityError:
		return slog.LevelError
	case SeverityCritical:
		return slog.LevelError + 4 // Higher than Error
	default:
		return slog.LevelInfo
	}
}

// ShouldCaptureStack returns true if stack traces should be captured for this severity
func (s Severity) ShouldCaptureStack() bool {
	return s == SeverityError || s == SeverityCritical
}
