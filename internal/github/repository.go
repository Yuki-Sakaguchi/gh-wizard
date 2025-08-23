package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/cli/go-gh/v2/pkg/api"
)

// RepositoryService provides GitHub repository operations
type RepositoryService struct {
	client *api.RESTClient
}

// NewRepositoryService creates a new repository service
func NewRepositoryService() (*RepositoryService, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, models.NewGitHubError(
			"Failed to initialize GitHub CLI",
			err,
		)
	}

	return &RepositoryService{client: client}, nil
}

// CreateRepository creates a GitHub repository
func (rs *RepositoryService) CreateRepository(ctx context.Context, config *models.ProjectConfig) (*RepositoryInfo, error) {
	if !config.CreateGitHub {
		return nil, nil // GitHub repository creation not required
	}

	// Get current user information
	user, err := rs.getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	// Check for repository duplication
	if err := rs.checkRepositoryExists(ctx, user.Login, config.Name); err != nil {
		return nil, err
	}

	// Create repository
	repoInfo, err := rs.createRepositoryViaAPI(ctx, config)
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

// getCurrentUser gets current GitHub user information
func (rs *RepositoryService) getCurrentUser(ctx context.Context) (*GitHubUser, error) {
	var user GitHubUser
	err := rs.client.Get("user", &user)
	if err != nil {
		return nil, models.NewGitHubError(
			"Failed to get user information",
			err,
		)
	}
	return &user, nil
}

// checkRepositoryExists checks for repository duplication
func (rs *RepositoryService) checkRepositoryExists(ctx context.Context, owner, name string) error {
	var repo RepositoryInfo
	err := rs.client.Get(fmt.Sprintf("repos/%s/%s", owner, name), &repo)

	if err == nil {
		// Repository exists
		return models.NewGitHubError(
			fmt.Sprintf("Repository '%s/%s' already exists", owner, name),
			nil,
		)
	}

	// 404 error is normal (repository doesn't exist)
	if strings.Contains(err.Error(), "404") {
		return nil
	}

	// Other errors
	return models.NewGitHubError(
		"Failed to check repository existence",
		err,
	)
}

// createRepositoryViaAPI creates repository via API
func (rs *RepositoryService) createRepositoryViaAPI(ctx context.Context, config *models.ProjectConfig) (*RepositoryInfo, error) {
	// Create request body
	createReq := CreateRepositoryRequest{
		Name:        config.Name,
		Description: config.Description,
		Private:     config.IsPrivate,
		AutoInit:    false, // Already initialized by template or locally
	}

	// For template repository
	if config.HasTemplate() {
		createReq.TemplateOwner = config.Template.Owner
		createReq.TemplateRepo = config.Template.Name
	}

	// Convert request body to JSON
	jsonData, err := json.Marshal(createReq)
	if err != nil {
		return nil, models.NewGitHubError(
			"Failed to create request data",
			err,
		)
	}

	var repoInfo RepositoryInfo
	err = rs.client.Post("user/repos", bytes.NewReader(jsonData), &repoInfo)
	if err != nil {
		return nil, models.NewGitHubError(
			fmt.Sprintf("Failed to create repository: %v", err),
			err,
		)
	}

	return &repoInfo, nil
}

// Data structures

// GitHubUser represents GitHub user information
type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// RepositoryInfo represents repository information
type RepositoryInfo struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	FullName  string     `json:"full_name"`
	Owner     GitHubUser `json:"owner"`
	Private   bool       `json:"private"`
	HTMLURL   string     `json:"html_url"`
	CloneURL  string     `json:"clone_url"`
	SSHURL    string     `json:"ssh_url"`
	GitURL    string     `json:"git_url"`
	CreatedAt string     `json:"created_at"`
}

// CreateRepositoryRequest represents repository creation request
type CreateRepositoryRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Private       bool   `json:"private"`
	AutoInit      bool   `json:"auto_init"`
	TemplateOwner string `json:"template_owner,omitempty"`
	TemplateRepo  string `json:"template_repo,omitempty"`
}
