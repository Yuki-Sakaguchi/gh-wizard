package models

import "fmt"

// ErrorType はエラーの種類を表す
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
	// システムエラー
	ErrGHNotInstalled     ErrorCode = "GH_NOT_INSTALLED"
	ErrGHNotAuthenticated ErrorCode = "GH_NOT_AUTHENTICATED"
	ErrAPIRateLimit       ErrorCode = "API_RATE_LIMIT"
	ErrNetworkError       ErrorCode = "NETWORK_ERROR"

	// ユーザーエラー
	ErrInvalidRepoName   ErrorCode = "INVALID_REPO_NAME"
	ErrRepoAlreadyExists ErrorCode = "REPO_ALREADY_EXISTS"
	ErrNoTemplatesFound  ErrorCode = "NO_TEMPLATES_FOUND"

	// アプリケーションエラー
	ErrConfigLoadFailed ErrorCode = "CONFIG_LOAD_FAILED"
	ErrTUIInitFailed    ErrorCode = "TUI_INIT_FAILED"
)

// WizardError は gh-wizard 固有のエラーを表す
type WizardError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// Error は error インターフェースを満たす
func (we *WizardError) Error() string {
	if we.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s", we.Type, we.Message, we.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", we.Type, we.Message)
}

// Unwrap は原因となったエラーを返す
func (we *WizardError) Unwrap() error {
	return we.Cause
}

// IsRetryable はリトライ可能なエラーかどうかを返す
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

// NewValidationError はバリデーションエラーを作成する
func NewValidationError(message string) *WizardError {
	return &WizardError{
		Type:    ErrorTypeValidation,
		Message: message,
		Cause:   nil,
	}
}

// NewGitHubError はGitHubエラーを作成する
func NewGitHubError(message string, cause error) *WizardError {
	return &WizardError{
		Type:    ErrorTypeGitHub,
		Message: message,
		Cause:   cause,
	}
}

// NewProjectError はプロジェクトエラーを作成する
func NewProjectError(message string, cause error) *WizardError {
	return &WizardError{
		Type:    ErrorTypeProject,
		Message: message,
		Cause:   cause,
	}
}

// NewWizardError は新しい WizardError を作成する (legacy)
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
