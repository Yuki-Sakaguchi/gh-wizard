package models

import "fmt"

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeGitHub     ErrorType = "GITHUB"
	ErrorTypeNetwork    ErrorType = "NETWORK"
	ErrorTypeProject    ErrorType = "PROJECT"
)


// WizardError represents gh-wizard specific errors
type WizardError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error satisfies the error interface
func (we *WizardError) Error() string {
	if we.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s", we.Type, we.Message, we.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", we.Type, we.Message)
}

// Unwrap returns the underlying error
func (we *WizardError) Unwrap() error {
	return we.Cause
}

// IsRetryable returns whether the error is retryable
func (we *WizardError) IsRetryable() bool {
	switch we.Type {
	case ErrorTypeNetwork, ErrorTypeGitHub:
		return true
	case ErrorTypeValidation, ErrorTypeProject:
		return false
	default:
		return false
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *WizardError {
	return &WizardError{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   nil,
	}
}

// NewGitHubError creates a GitHub error
func NewGitHubError(message string, cause error) *WizardError {
	return &WizardError{
		Type:    ErrorTypeGitHub,
		Message: message,
		Cause:   cause,
	}
}

// NewProjectError creates a project error
func NewProjectError(message string, cause error) *WizardError {
	return &WizardError{
		Type:    ErrorTypeProject,
		Message: message,
		Cause:   cause,
	}
}

