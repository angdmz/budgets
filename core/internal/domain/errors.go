package domain

import (
	"errors"
	"fmt"
)

// wrapNotFound annotates ErrNotFound with a resource-specific message while
// letting any other persistence error propagate unmasked.
func wrapNotFound(err error, message string) error {
	if errors.Is(err, ErrNotFound) {
		return fmt.Errorf("%w: %s", ErrNotFound, message)
	}
	return err
}

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("access forbidden")
	ErrConflict          = errors.New("resource conflict")
	ErrGone              = errors.New("resource no longer available")
	ErrValidation        = errors.New("validation error")
	ErrInternalServer    = errors.New("internal server error")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	return v.Errors[0].Message
}
