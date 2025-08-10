package models

import "fmt"

// ErrorCode はエラーコードを表す
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
	Code    ErrorCode
	Message string
	Cause   error
	Step    Step
	Retry   bool
}

// Error は error インターフェースを満たす
func (we WizardError) Error() string {
	if we.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", we.Code, we.Message, we.Cause)
	}
	return fmt.Sprintf("[%s] %s", we.Code, we.Message)
}

// NewWizardError は新しい WizardError を作成する
func NewWizardError(code ErrorCode, message string, cause error) *WizardError {
	return &WizardError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Retry:   false,
	}
}
