package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"

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

// SearchPopularTemplates は人気のテンプレートリポジトリを検索する
func (c *DefaultClient) SearchPopularTemplates(ctx context.Context) ([]models.Template, error) {
	// GitHub APIでテンプレートリポジトリを検索
	queries := []string{
		"template react sort:stars",
		"template vue sort:stars", 
		"template nextjs sort:stars",
		"template golang sort:stars",
		"template python sort:stars",
		"template typescript sort:stars",
	}
	
	var allTemplates []models.Template
	
	for _, query := range queries {
		templates, err := c.searchRepositories(ctx, query, 3) // 各カテゴリから3つずつ
		if err != nil {
			// エラーログを出力するが、処理は継続
			fmt.Printf("Warning: failed to search templates for query '%s': %v\n", query, err)
			continue
		}
		allTemplates = append(allTemplates, templates...)
	}
	
	// 重複を除去し、スター数でソート
	uniqueTemplates := removeDuplicateTemplates(allTemplates)
	sort.Slice(uniqueTemplates, func(i, j int) bool {
		return uniqueTemplates[i].Stars > uniqueTemplates[j].Stars
	})
	
	// 上位20個に制限
	if len(uniqueTemplates) > 20 {
		uniqueTemplates = uniqueTemplates[:20]
	}
	
	return uniqueTemplates, nil
}

// searchRepositories はGitHub CLIを使ってリポジトリを検索する
func (c *DefaultClient) searchRepositories(ctx context.Context, query string, limit int) ([]models.Template, error) {
	cmd := exec.CommandContext(ctx, "gh", "search", "repos", query, "--limit", strconv.Itoa(limit), "--json", "name,owner,stargazersCount,description")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}
	
	var searchResults []struct {
		Name           string `json:"name"`
		Owner          struct {
			Login string `json:"login"`
		} `json:"owner"`
		StargazersCount int    `json:"stargazersCount"`
		Description     string `json:"description"`
	}
	
	if err := json.Unmarshal(output, &searchResults); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}
	
	templates := make([]models.Template, len(searchResults))
	for i, result := range searchResults {
		templates[i] = models.Template{
			Name:        result.Name,
			FullName:    fmt.Sprintf("%s/%s", result.Owner.Login, result.Name),
			Stars:       result.StargazersCount,
			Description: result.Description,
		}
	}
	
	return templates, nil
}

// removeDuplicateTemplates は重複するテンプレートを除去する
func removeDuplicateTemplates(templates []models.Template) []models.Template {
	seen := make(map[string]bool)
	var unique []models.Template
	
	for _, template := range templates {
		if !seen[template.FullName] {
			seen[template.FullName] = true
			unique = append(unique, template)
		}
	}
	
	return unique
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
