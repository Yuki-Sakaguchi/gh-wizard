package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWizardError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *WizardError
		expected string
	}{
		{
			name: "error with cause",
			err: &WizardError{
				Type:    ErrorTypeGitHub,
				Message: "API failed",
				Cause:   errors.New("network error"),
			},
			expected: "[GITHUB] API failed: network error",
		},
		{
			name: "error without cause",
			err: &WizardError{
				Type:    ErrorTypeValidation,
				Message: "Invalid input",
				Cause:   nil,
			},
			expected: "[VALIDATION] Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWizardError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wizardErr := &WizardError{
		Type:    ErrorTypeNetwork,
		Message: "Network failed",
		Cause:   originalErr,
	}

	unwrapped := wizardErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)
}

func TestWizardError_IsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "network error is retryable",
			errorType: ErrorTypeNetwork,
			expected:  true,
		},
		{
			name:      "github error is retryable",
			errorType: ErrorTypeGitHub,
			expected:  true,
		},
		{
			name:      "validation error is not retryable",
			errorType: ErrorTypeValidation,
			expected:  false,
		},
		{
			name:      "project error is not retryable",
			errorType: ErrorTypeProject,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &WizardError{Type: tt.errorType}
			result := err.IsRetryable()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWizardError_Constructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("invalid input")

		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "invalid input", err.Message)
		assert.Nil(t, err.Cause)
	})

	t.Run("NewGitHubError", func(t *testing.T) {
		cause := errors.New("api error")
		err := NewGitHubError("github failed", cause)

		assert.Equal(t, ErrorTypeGitHub, err.Type)
		assert.Equal(t, "github failed", err.Message)
		assert.Equal(t, cause, err.Cause)
	})
}
