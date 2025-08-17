package github

import (
	"context"
	"testing"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
)

// SimpleMockClient は設定可能なモッククライアント
type SimpleMockClient struct {
	Templates     []models.Template
	AuthError     error
	CreateError   error
	TemplateError error
}

// NewSimpleMockClient は新しいシンプルモッククライアントを作成する
func NewSimpleMockClient() *SimpleMockClient {
	return &SimpleMockClient{
		Templates: []models.Template{
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
		},
	}
}

// SearchPopularTemplates はモックテンプレート検索
func (c *SimpleMockClient) SearchPopularTemplates(ctx context.Context) ([]models.Template, error) {
	if c.TemplateError != nil {
		return nil, c.TemplateError
	}
	return c.Templates, nil
}

// CheckAuthentication はモック認証チェック
func (m *SimpleMockClient) CheckAuthentication(ctx context.Context) error {
	return m.AuthError
}

// GetUserTemplates はモックテンプレート取得
func (m *SimpleMockClient) GetUserTemplates(ctx context.Context) ([]models.Template, error) {
	if m.TemplateError != nil {
		return nil, m.TemplateError
	}
	if m.AuthError != nil {
		return nil, m.AuthError
	}
	return m.Templates, nil
}

// CreateRepository はモックリポジトリ作成
func (m *SimpleMockClient) CreateRepository(ctx context.Context, config *models.ProjectConfig) error {
	return m.CreateError
}

func TestSimpleMockClient_Scenarios(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*SimpleMockClient)
		expectError     bool
		expectTemplates int
	}{
		{
			name: "successful template retrieval",
			setupMock: func(m *SimpleMockClient) {
				// デフォルト設定（エラーなし）
			},
			expectError:     false,
			expectTemplates: 2,
		},
		{
			name: "authentication error",
			setupMock: func(m *SimpleMockClient) {
				m.AuthError = models.NewGitHubError("auth failed", nil)
			},
			expectError:     true,
			expectTemplates: 0,
		},
		{
			name: "template error",
			setupMock: func(m *SimpleMockClient) {
				m.TemplateError = models.NewWizardError(models.ErrNoTemplatesFound, "no templates", nil)
			},
			expectError:     true,
			expectTemplates: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewSimpleMockClient()
			tt.setupMock(mockClient)

			ctx := context.Background()
			templates, err := mockClient.GetUserTemplates(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, templates, tt.expectTemplates)
		})
	}
}
