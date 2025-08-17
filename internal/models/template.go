package models

import (
	"fmt"
	"time"
)

// Template は GitHub のテンプレートリポジトリの情報を表す
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Owner       string    `json:"owner"`
	Description string    `json:"description"`
	Stars       int       `json:"stars"`
	Forks       int       `json:"forks"`
	Language    string    `json:"language"`
	Topics      []string  `json:"topics"`
	IsTemplate  bool      `json:"is_template"`
	Private     bool      `json:"private"`
	UpdatedAt   time.Time `json:"updated_at"`
	CloneURL    string    `json:"clone_url"`
}

// GetDisplayName はテンプレートリポジトリの表示名を返す
func (t Template) GetDisplayName() string {
	result := t.Name
	
	// Stars情報を追加
	if t.Stars > 0 {
		result += fmt.Sprintf(" (⭐ %d)", t.Stars)
	}
	
	// 言語情報を追加
	if t.Language != "" {
		result += fmt.Sprintf(" [%s]", t.Language)
	}
	
	return result
}

// GetShortDescription は短縮された説明を返す
func (t Template) GetShortDescription() string {
	if t.Description == "" {
		return "説明なし"
	}
	if len(t.Description) > 72 {
		return t.Description[:72] + "..."
	}
	return t.Description
}

// GetRepoURL はリポジトリのURLを返す
func (t Template) GetRepoURL() string {
	return "https://github.com/" + t.FullName
}

// GetIsPublic はパブリックリポジトリかどうかを返す
func (t Template) GetIsPublic() bool {
	return !t.Private
}
