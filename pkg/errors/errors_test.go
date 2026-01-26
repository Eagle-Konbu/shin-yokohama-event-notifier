package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainError_Error_WithWrappedError(t *testing.T) {
	wrappedErr := errors.New("underlying error")
	domainErr := &DomainError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Err:     wrappedErr,
	}

	errorStr := domainErr.Error()

	assert.Contains(t, errorStr, "TEST_ERROR")
	assert.Contains(t, errorStr, "test message")
	assert.Contains(t, errorStr, "underlying error")
}

func TestDomainError_Error_WithoutWrappedError(t *testing.T) {
	domainErr := &DomainError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Err:     nil,
	}

	errorStr := domainErr.Error()

	assert.Contains(t, errorStr, "TEST_ERROR")
	assert.Contains(t, errorStr, "test message")
	assert.NotContains(t, errorStr, "<nil>")
}

func TestDomainError_Unwrap_WithWrappedError(t *testing.T) {
	wrappedErr := errors.New("underlying error")
	domainErr := &DomainError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Err:     wrappedErr,
	}

	unwrapped := domainErr.Unwrap()

	assert.Equal(t, wrappedErr, unwrapped)
	assert.ErrorIs(t, domainErr, wrappedErr)
}

func TestDomainError_Unwrap_WithoutWrappedError(t *testing.T) {
	domainErr := &DomainError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Err:     nil,
	}

	unwrapped := domainErr.Unwrap()

	assert.Nil(t, unwrapped)
}

func TestNewInfrastructureError(t *testing.T) {
	wrappedErr := errors.New("underlying error")
	message := "infrastructure failure"

	domainErr := NewInfrastructureError(message, wrappedErr)

	require.NotNil(t, domainErr)
	assert.Equal(t, "INFRASTRUCTURE_ERROR", domainErr.Code)
	assert.Equal(t, message, domainErr.Message)
	assert.Equal(t, wrappedErr, domainErr.Err)
	assert.ErrorIs(t, domainErr, wrappedErr)
}

func TestNewInfrastructureError_NilWrappedError(t *testing.T) {
	message := "infrastructure failure"

	domainErr := NewInfrastructureError(message, nil)

	require.NotNil(t, domainErr)
	assert.Equal(t, "INFRASTRUCTURE_ERROR", domainErr.Code)
	assert.Equal(t, message, domainErr.Message)
	assert.Nil(t, domainErr.Err)
}

func TestNewValidationError(t *testing.T) {
	message := "validation failed"

	domainErr := NewValidationError(message)

	require.NotNil(t, domainErr)
	assert.Equal(t, "VALIDATION_ERROR", domainErr.Code)
	assert.Equal(t, message, domainErr.Message)
	assert.Nil(t, domainErr.Err)
}

func TestDomainError_ErrorIs_Integration(t *testing.T) {
	originalErr := errors.New("original error")
	infraErr := NewInfrastructureError("infrastructure issue", originalErr)

	assert.True(t, errors.Is(infraErr, originalErr))

	var domainErr *DomainError
	assert.True(t, errors.As(infraErr, &domainErr))
	assert.Equal(t, "INFRASTRUCTURE_ERROR", domainErr.Code)
}

func TestDomainError_ErrorChaining(t *testing.T) {
	baseErr := errors.New("base error")
	infraErr := NewInfrastructureError("layer 1", baseErr)
	validationErr := NewValidationError("layer 2")

	assert.ErrorIs(t, infraErr, baseErr)
	assert.Contains(t, infraErr.Error(), "INFRASTRUCTURE_ERROR")
	assert.Contains(t, infraErr.Error(), "layer 1")
	assert.Contains(t, infraErr.Error(), "base error")

	assert.NotErrorIs(t, validationErr, baseErr)
	assert.Contains(t, validationErr.Error(), "VALIDATION_ERROR")
	assert.Contains(t, validationErr.Error(), "layer 2")
}
