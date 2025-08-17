package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClient はテスト用のモッククライアント
type MockClient struct {
	mock.Mock
}

func (m *MockClient) CheckAuthentication(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetUserTemplates(ctx context.Context) ([]models.Template, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Template), args.Error(1)
}

func (m *MockClient) CreateRepository(ctx context.Context, config *models.ProjectConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func TestMockClient_GetUserTemplates_Success(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()

	expectedTemplates := []models.Template{
		{
			ID:          "1",
			Name:        "nextjs-starter",
			FullName:    "testuser/nextjs-starter",
			Owner:       "testuser",
			Description: "Next.js starter template",
			Stars:       15,
			Language:    "TypeScript",
			IsTemplate:  true,
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "2",
			Name:        "go-cli-template",
			FullName:    "testuser/go-cli-template",
			Owner:       "testuser",
			Description: "Go CLI application template",
			Stars:       8,
			Language:    "Go",
			IsTemplate:  true,
			UpdatedAt:   time.Now(),
		},
	}

	mockClient.On("GetUserTemplates", ctx).Return(expectedTemplates, nil)

	templates, err := mockClient.GetUserTemplates(ctx)

	require.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Equal(t, "nextjs-starter", templates[0].Name)
	assert.Equal(t, 15, templates[0].Stars)
	assert.Equal(t, "TypeScript", templates[0].Language)

	mockClient.AssertExpectations(t)
}

func TestMockCilent_GetUserTemplates_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError string
	}{
		{
			name:          "network error",
			mockError:     models.NewWizardError(models.ErrNetworkError, "connection failed", errors.New("dial failed")),
			expectedError: "connection failed",
		},
		{
			name:          "authentication error",
			mockError:     models.NewGitHubError("authentication", errors.New("401")),
			expectedError: "authentication failed",
		},
		{
			name:          "no templates found",
			mockError:     models.NewWizardError(models.ErrNoTemplatesFound, "no templates found", nil),
			expectedError: "no templates found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			ctx := context.Background()

			mockClient.On("GetUserTemplates", ctx).Return([]models.Template{}, tt.mockError)

			template, err := mockClient.GetUserTemplates(ctx)

			require.Error(t, err)
			assert.Empty(t, template)
			assert.Contains(t, err.Error(), tt.expectedError)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockClient_CheckAuthentication(t *testing.T) {
	tests := []struct {
		name      string
		mockError error
		wantErr   bool
	}{
		{
			name:      "authentication success",
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "authentication failure",
			mockError: models.NewGitHubError("authentication failed", errors.New("401")),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockClient)
			ctx := context.Background()

			mockClient.On("CheckAuthentication", ctx).Return(tt.mockError)

			err := mockClient.CheckAuthentication(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestMockClient_CreateRepository(t *testing.T) {
	mockClient := new(MockClient)
	ctx := context.Background()

	config := &models.ProjectConfig{
		Name:         "test-project",
		Description:  "Test project",
		CreateGitHub: true,
		IsPrivate:    true,
	}

	mockClient.On("CreateRepository", ctx, config).Return(nil)

	err := mockClient.CreateRepository(ctx, config)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
