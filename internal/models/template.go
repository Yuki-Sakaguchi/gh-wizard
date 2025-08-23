package models

import (
	"fmt"
	"time"
)

// Template represents GitHub template repository information
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

// GetDisplayName returns the display name of template repository
func (t Template) GetDisplayName() string {
	result := t.Name

	// Add Stars information
	if t.Stars > 0 {
		result += fmt.Sprintf(" (⭐ %d)", t.Stars)
	}

	// Add language information
	if t.Language != "" {
		result += fmt.Sprintf(" [%s]", t.Language)
	}

	return result
}

// GetShortDescription returns shortened description
func (t Template) GetShortDescription() string {
	if t.Description == "" {
		return "No description"
	}
	if len(t.Description) > 72 {
		return t.Description[:72] + "..."
	}
	return t.Description
}

// GetRepoURL returns repository URL
func (t Template) GetRepoURL() string {
	return "https://github.com/" + t.FullName
}

// GetIsPublic returns whether the repository is public
func (t Template) GetIsPublic() bool {
	return !t.Private
}
