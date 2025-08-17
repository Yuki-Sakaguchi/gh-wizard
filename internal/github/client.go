package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

type Client interface {
	// GetUserTemplate はユーザーのテンプレートリポジトリを取得する
	GetUserTemplates(ctx context.Context) ([]models.Template, error)

	// SearchPopularTemplates は人気のテンプレートリポジトリを検索する
	SearchPopularTemplates(ctx context.Context) ([]models.Template, error)

	// CreateRepository は GitHub リポジトリを作成する
	CreateRepository(ctx context.Context, config *models.ProjectConfig) error

	// CheckAuthentication は 認証状態を確認する
	CheckAuthentication(ctx context.Context) error
}

// DefaultClient は go-gh で使用するデフォルト実装
type DefaultClient struct {
	// go-gh クライアントは内部で管理
}

// NewClient は新しい GitHub クライアントを作成する
func NewClient() Client {
	return &DefaultClient{}
}

// GetUserTemplates はユーザーのテンプレートリポジトリを取得する
func (c *DefaultClient) GetUserTemplates(ctx context.Context) ([]models.Template, error) {
	// TODO: Issue #28で実装予定
	// テスト用に空のスライスを返す
	return []models.Template{}, nil
}

// SearchPopularTemplates はユーザー自身のテンプレートリポジトリを取得する
func (c *DefaultClient) SearchPopularTemplates(ctx context.Context) ([]models.Template, error) {
	// 認証されたユーザーのリポジトリのみを取得
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", "--json", "name,owner,stargazerCount,description,isTemplate", "--limit", "100")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to get user repositories: %w", err)
	}
	
	var repositories []struct {
		Name           string `json:"name"`
		Owner          struct {
			Login string `json:"login"`
		} `json:"owner"`
		StargazerCount int    `json:"stargazerCount"`
		Description    string `json:"description"`
		IsTemplate     bool   `json:"isTemplate"`
	}
	
	if err := json.Unmarshal(output, &repositories); err != nil {
		return nil, fmt.Errorf("failed to parse repository list: %w", err)
	}
	
	// テンプレートリポジトリのみをフィルタ
	var templates []models.Template
	for _, repo := range repositories {
		if repo.IsTemplate {
			templates = append(templates, models.Template{
				Name:        repo.Name,
				FullName:    fmt.Sprintf("%s/%s", repo.Owner.Login, repo.Name),
				Stars:       repo.StargazerCount,
				Description: repo.Description,
			})
		}
	}
	
	// スター数でソート
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Stars > templates[j].Stars
	})
	
	return templates, nil
}


// CreateRepository は GitHub リポジトリを作成する
func (c *DefaultClient) CreateRepository(ctx context.Context, config *models.ProjectConfig) error {
	// TODO: Issue #28で実装予定
	return nil
}

// CheckAuthentication は認証状態を確認する
func (c *DefaultClient) CheckAuthentication(ctx context.Context) error {
	// TODO: Issue #28で実装予定
	return nil
}

// GetTemplateByFullName は完全名でテンプレートを検索する
func GetTemplateByFullName(templates []models.Template, fullName string) *models.Template {
	if fullName == "" {
		return nil
	}

	for _, template := range templates {
		if template.FullName == fullName {
			return &template
		}
	}
	return nil
}

// GetTemplateByDisplayName は表示名でテンプレートを検索する
func GetTemplateByDisplayName(templates []models.Template, displayName string) *models.Template {
	for _, template := range templates {
		if template.GetDisplayName() == displayName {
			return &template
		}
	}
	return nil
}

// SortTemplatesByStars はスター数でテンプレートをソートする
func SortTemplatesByStars(templates []models.Template) {
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Stars > templates[j].Stars
	})
}

// SortTemplatesByUpdated は更新日時でテンプレートをソートする
func SortTemplatesByUpdated(templates []models.Template) {
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].UpdatedAt.After(templates[j].UpdatedAt)
	})
}
