package validation

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	hasUpperRegex   = regexp.MustCompile(`[A-Z]`)
	hasLowerRegex   = regexp.MustCompile(`[a-z]`)
	hasDigitRegex   = regexp.MustCompile(`\d`)
	hasSpecialRegex = regexp.MustCompile(`[@$!%*?&]`)
)

// RegisterDefaultCustomValidations registers all default custom validators
func (s *Service) RegisterDefaultCustomValidations() error {
	if err := s.RegisterCustomValidation("strongpassword", validateStrongPassword); err != nil {
		return err
	}

	if err := s.RegisterCustomValidation("filepath", validateFilePath); err != nil {
		return err
	}

	return nil
}

// validateFilePath validates that a file path is safe (no traversal, no backslash, no absolute paths)
func validateFilePath(fl validator.FieldLevel) bool {
	path := fl.Field().String()
	if strings.Contains(path, "..") || strings.Contains(path, "\\") {
		return false
	}
	cleaned := filepath.Clean(path)
	return !filepath.IsAbs(cleaned)
}

// validateStrongPassword validates that a password meets strong password requirements
// Password must contain:
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - At least one special character (@$!%*?&)
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Minimum length check (should also use min tag, but double-check here)
	if len(password) < 8 {
		return false
	}

	// Check for required character types
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case char == '@' || char == '$' || char == '!' || char == '%' || char == '*' || char == '?' || char == '&':
			hasSpecial = true
		}

		// Early exit if all requirements are met
		if hasUpper && hasLower && hasDigit && hasSpecial {
			return true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}
