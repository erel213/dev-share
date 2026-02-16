package validation

import (
	"reflect"
	"strings"

	pkgerrors "backend/pkg/errors"

	"github.com/go-playground/validator/v10"
)

// Service wraps the go-playground validator for domain use
type Service struct {
	validate *validator.Validate
}

// New creates a new validation service
func New() *Service {
	v := validator.New()

	// Use JSON field names instead of struct field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Service{validate: v}
}

// Validate validates a struct and returns a domain error if validation fails
func (s *Service) Validate(data interface{}) *pkgerrors.Error {
	err := s.validate.Struct(data)
	if err == nil {
		return nil
	}

	// Handle validation errors
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// Unexpected error type
		return pkgerrors.WithCode(
			pkgerrors.CodeValidation,
			"validation failed: "+err.Error(),
		).WithHTTPStatus(400)
	}

	// Convert to field error map
	fieldErrors := make(map[string]string)
	for _, fieldErr := range validationErrs {
		fieldName := fieldErr.Field()
		fieldErrors[fieldName] = formatValidationError(fieldErr)
	}

	// Use domain ValidationError helper
	return pkgerrors.WithCode(
		pkgerrors.CodeValidation,
		"validation failed",
	).
		WithHTTPStatus(400).
		WithSeverity(pkgerrors.SeverityWarning).
		WithMetadata("fields", fieldErrors)
}

// RegisterCustomValidation registers a custom validation function
func (s *Service) RegisterCustomValidation(tag string, fn validator.Func) error {
	return s.validate.RegisterValidation(tag, fn)
}

// formatValidationError converts a validator.FieldError to a human-readable message
func formatValidationError(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		if fe.Kind() == reflect.String {
			return field + " must be at least " + fe.Param() + " characters"
		}
		return field + " must be at least " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return field + " must be at most " + fe.Param() + " characters"
		}
		return field + " must be at most " + fe.Param()
	case "len":
		return field + " must be exactly " + fe.Param() + " characters"
	case "uuid4":
		return field + " must be a valid UUID v4"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gt":
		return field + " must be greater than " + fe.Param()
	case "gte":
		return field + " must be greater than or equal to " + fe.Param()
	case "lt":
		return field + " must be less than " + fe.Param()
	case "lte":
		return field + " must be less than or equal to " + fe.Param()
	case "eq":
		return field + " must equal " + fe.Param()
	case "ne":
		return field + " must not equal " + fe.Param()
	case "strongpassword":
		return field + " must contain at least one uppercase letter, one lowercase letter, one number, and one special character"
	default:
		// Fallback for unknown tags
		return field + " failed validation: " + fe.Tag()
	}
}
