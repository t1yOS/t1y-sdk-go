package t1y

import "fmt"

// T1YError is a custom error type for t1yOS SDK errors.
// It wraps API error responses with code, message, and data.
type T1YError struct {
	Code    int
	Message string
	Data    any
}

// Error implements the error interface.
func (e *T1YError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("T1YError [%d]: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("T1YError [%d]: %s", e.Code, e.Message)
}

// NewT1YError creates a new T1YError.
func NewT1YError(code int, message string, data any) *T1YError {
	return &T1YError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// ValidationError is a validation error for invalid configuration parameters.
type ValidationError struct {
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("ValidationError: %s", e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}
