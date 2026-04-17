package domain

import "fmt"

// ErrNotFound is returned when a config does not exist in the database.
// We define it here so handler can check the error type without
// importing the repository package — layers stay decoupled.
var ErrNotFound = fmt.Errorf("config not found")

// ValidationError holds a human-readable message for bad input.
// Using a named type lets handlers do: errors.As(err, &ValidationError{})
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ErrMissingField returns a ValidationError for a missing required field.
func ErrMissingField(field string) error {
	return &ValidationError{Message: fmt.Sprintf("field '%s' is required", field)}
}

// ErrInvalidField returns a ValidationError for a field that has a bad value.
func ErrInvalidField(field, reason string) error {
	return &ValidationError{Message: fmt.Sprintf("field '%s' is invalid: %s", field, reason)}
}
