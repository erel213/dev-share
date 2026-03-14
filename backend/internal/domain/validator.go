package domain

import pkgerrors "backend/pkg/errors"

// Validator validates structs and returns a domain error if validation fails.
// Implementations may use struct tags (e.g. go-playground/validator) or custom logic.
type Validator interface {
	Validate(data interface{}) *pkgerrors.Error
}
