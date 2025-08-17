package github

import (
	"context"
	"sort"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

type Client interface {
	// GetUserTemplate はユーザーのテンプレートリポジトリを取得する
	GetUserTemplates(ctx context.Context) ([]models.Template, error)

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
	return &DefaultClient()
}

// GetTemplateByFullName は完全名でテンプレートを検索する
func GetTemplateByFullName(templates []models.Template, fullName string) *models.Template {
	if fullName == "" {
		return nil
	}

	for _, tempalate := range templates {
		if tempalate.FullName == fullName {
			return &tempalate
		}
	}
	return nil
}

// GetTemplateByDisplayName は表示名でテンプレートを検索する
func GetTemplateByDisplayName(templates []models.Template, displayName string) *models.Template {
	for _, template := range templates {
		if template.GetDisplayName()() == displayName {
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
