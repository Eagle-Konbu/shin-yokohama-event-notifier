package errors

import "fmt"

type DomainError struct {
	Err     error
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

func NewInfrastructureError(message string, err error) *DomainError {
	return &DomainError{
		Code:    "INFRASTRUCTURE_ERROR",
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Code:    "VALIDATION_ERROR",
		Message: message,
	}
}
