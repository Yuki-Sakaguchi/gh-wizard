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
	// GetUserTemplates gets user's template repositories
	GetUserTemplates(ctx context.Context) ([]models.Template, error)

	// SearchPopularTemplates searches for popular template repositories
	SearchPopularTemplates(ctx context.Context) ([]models.Template, error)

	// CreateRepository creates a GitHub repository
	CreateRepository(ctx context.Context, config *models.ProjectConfig) error

	// CheckAuthentication checks authentication status
	CheckAuthentication(ctx context.Context) error
}

// DefaultClient is the default implementation using go-gh
type DefaultClient struct {
	// go-gh client is managed internally
}

// NewClient creates a new GitHub client
func NewClient() Client {
	return &DefaultClient{}
}

// GetUserTemplates gets user's template repositories
func (c *DefaultClient) GetUserTemplates(ctx context.Context) ([]models.Template, error) {
	// TODO: Implementation planned in Issue #28
	// Return empty slice for testing
	return []models.Template{}, nil
}

// SearchPopularTemplates gets user's own template repositories
func (c *DefaultClient) SearchPopularTemplates(ctx context.Context) ([]models.Template, error) {
	// Get only authenticated user's repositories
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", "--json", "name,owner,stargazerCount,description,isTemplate", "--limit", "100")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to get user repositories: %w", err)
	}

	var repositories []struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		StargazerCount int    `json:"stargazerCount"`
		Description    string `json:"description"`
		IsTemplate     bool   `json:"isTemplate"`
	}

	if err := json.Unmarshal(output, &repositories); err != nil {
		return nil, fmt.Errorf("failed to parse repository list: %w", err)
	}

	// Filter only template repositories
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

	// Sort by star count
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Stars > templates[j].Stars
	})

	return templates, nil
}

// CreateRepository creates a GitHub repository
func (c *DefaultClient) CreateRepository(ctx context.Context, config *models.ProjectConfig) error {
	// TODO: Implementation planned in Issue #28
	return nil
}

// CheckAuthentication checks authentication status
func (c *DefaultClient) CheckAuthentication(ctx context.Context) error {
	// TODO: Implementation planned in Issue #28
	return nil
}

// GetTemplateByFullName searches for template by full name
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

// GetTemplateByDisplayName searches for template by display name
func GetTemplateByDisplayName(templates []models.Template, displayName string) *models.Template {
	for _, template := range templates {
		if template.GetDisplayName() == displayName {
			return &template
		}
	}
	return nil
}

// SortTemplatesByStars sorts templates by star count
func SortTemplatesByStars(templates []models.Template) {
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Stars > templates[j].Stars
	})
}

// SortTemplatesByUpdated sorts templates by updated date
func SortTemplatesByUpdated(templates []models.Template) {
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].UpdatedAt.After(templates[j].UpdatedAt)
	})
}
