package models

import (
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
}

// GetDisplayName はテンプレートリポジトリの表示名を返す
func (t Template) GetDisplayName() string {
	if t.Description != "" {
		return t.Name + " - " + t.Description
	}
	return t.Name
}
