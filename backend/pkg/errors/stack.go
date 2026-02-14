package errors

import (
	"runtime"
	"strings"
)

// Frame represents a single stack frame
type Frame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// captureStack captures the stack trace starting from the given skip level
// skip is the number of stack frames to skip (typically 2-3 to skip error creation frames)
func captureStack(skip int) []uintptr {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])
	return pcs[:n]
}

// formatStack converts program counters to formatted stack frames
func formatStack(stack []uintptr) []Frame {
	if len(stack) == 0 {
		return nil
	}

	frames := make([]Frame, 0, len(stack))
	callersFrames := runtime.CallersFrames(stack)

	for {
		frame, more := callersFrames.Next()

		// Skip runtime and error package frames for cleaner output
		if !shouldSkipFrame(frame.Function) {
			frames = append(frames, Frame{
				File:     frame.File,
				Line:     frame.Line,
				Function: simplifyFunctionName(frame.Function),
			})
		}

		if !more {
			break
		}
	}

	return frames
}

// shouldSkipFrame determines if a frame should be skipped in the stack trace
func shouldSkipFrame(function string) bool {
	// Skip runtime internal functions
	if strings.HasPrefix(function, "runtime.") {
		return true
	}
	// Skip error package internal functions
	if strings.Contains(function, "backend/pkg/errors.") {
		return true
	}
	return false
}

// simplifyFunctionName extracts the package and function name from a full function path
// Example: "github.com/user/project/internal/service.(*Service).Method" -> "service.(*Service).Method"
func simplifyFunctionName(fullName string) string {
	// Find the last package separator
	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash == -1 {
		return fullName
	}

	// Extract everything after the last slash
	return fullName[lastSlash+1:]
}
