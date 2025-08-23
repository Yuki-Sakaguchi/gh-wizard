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

// Legacy ErrorCode constants for backward compatibility
type ErrorCode string

const (
	// System errors
	ErrGHNotInstalled     ErrorCode = "GH_NOT_INSTALLED"
	ErrGHNotAuthenticated ErrorCode = "GH_NOT_AUTHENTICATED"
	ErrAPIRateLimit       ErrorCode = "API_RATE_LIMIT"
	ErrNetworkError       ErrorCode = "NETWORK_ERROR"

	// User errors
	ErrInvalidRepoName   ErrorCode = "INVALID_REPO_NAME"
	ErrRepoAlreadyExists ErrorCode = "REPO_ALREADY_EXISTS"
	ErrNoTemplatesFound  ErrorCode = "NO_TEMPLATES_FOUND"

	// Application errors
	ErrConfigLoadFailed ErrorCode = "CONFIG_LOAD_FAILED"
	ErrTUIInitFailed    ErrorCode = "TUI_INIT_FAILED"
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

// NewWizardError creates a new WizardError (legacy)
func NewWizardError(code ErrorCode, message string, cause error) *WizardError {
	var errorType ErrorType
	switch code {
	case ErrNetworkError:
		errorType = ErrorTypeNetwork
	case ErrAPIRateLimit:
		errorType = ErrorTypeGitHub
	case ErrInvalidRepoName:
		errorType = ErrorTypeValidation
	default:
		errorType = ErrorTypeProject
	}

	return &WizardError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}
