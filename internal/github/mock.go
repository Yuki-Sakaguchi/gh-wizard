package github

import (
	"context"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// MockClient はテスト用のモッククライアント
type MockClient struct {
	IsAuthenticatedFunc         func() error
	GetCurrentUserFunc          func() (*User, error)
	GetTemplateRepositoriesFunc func(ctx context.Context) ([]models.Template, error)
	CreateRepositoryFunc        func(ctx context.Context, config models.RepositoryConfig, template *models.Template) error
}

// IsAuthenticated はモック実装
func (m *MockClient) IsAuthenticated() error {
	if m.IsAuthenticatedFunc != nil {
		return m.IsAuthenticatedFunc()
	}
	return nil
}

// GetCurrentUser はモック実装
func (m *MockClient) GetCurrentUser() (*User, error) {
	if m.GetCurrentUserFunc != nil {
		return m.GetCurrentUserFunc()
	}
	return &User{
		Login: "testuser",
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}

// GetTemplateRepositories はモック実装
func (m *MockClient) GetTemplateRepositories(ctx context.Context) ([]models.Template, error) {
	if m.GetTemplateRepositoriesFunc != nil {
		return m.GetTemplateRepositoriesFunc(ctx)
	}

	// デフォルトのモックデータ
	return []models.Template{
		{
			ID:          "1",
			Name:        "nextjs-starter",
			FullName:    "testuser/nextjs-starter",
			Owner:       "testuser",
			Description: "Next.js + TypeScript starter template",
			Stars:       15,
			Forks:       3,
			Language:    "TypeScript",
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
			IsTemplate:  true,
			Private:     false,
		},
		{
			ID:          "2",
			Name:        "go-cli-template",
			FullName:    "testuser/go-cli-template",
			Owner:       "testuser",
			Description: "Go CLI application template",
			Stars:       8,
			Forks:       2,
			Language:    "Go",
			UpdatedAt:   time.Now().Add(-48 * time.Hour),
			IsTemplate:  true,
			Private:     false,
		},
	}, nil
}

// CreateRepository はモック実装
func (m *MockClient) CreateRepository(ctx context.Context, config models.RepositoryConfig, template *models.Template) error {
	if m.CreateRepositoryFunc != nil {
		return m.CreateRepositoryFunc(ctx, config, template)
	}
	return nil
}
