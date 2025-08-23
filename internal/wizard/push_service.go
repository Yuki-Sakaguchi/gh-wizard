package wizard

import (
	"context"
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/utils"
)

// PushService manages pushes to GitHub
type PushService struct {
	repoService *github.RepositoryService
	gitService  *utils.GitService
}

// NewPushService creates a new push service
func NewPushService(projectPath string) (*PushService, error) {
	repoService, err := github.NewRepositoryService()
	if err != nil {
		return nil, err
	}

	gitService := utils.NewGitService(projectPath)

	return &PushService{
		repoService: repoService,
		gitService:  gitService,
	}, nil
}

// PushToGitHub pushes local repository to GitHub
func (ps *PushService) PushToGitHub(ctx context.Context, config *models.ProjectConfig) error {
	if !config.CreateGitHub {
		return nil // GitHub push not required
	}

	// Create GitHub repository
	repoInfo, err := ps.repoService.CreateRepository(ctx, config)
	if err != nil {
		return err
	}

	if repoInfo == nil {
		return nil // Repository creation skipped
	}

	// Initialize Git and commit
	if err := ps.initializeLocalRepository(ctx, config); err != nil {
		return err
	}

	// Configure remote repository and push
	if err := ps.pushToRemoteRepository(ctx, repoInfo); err != nil {
		return err
	}

	// Success message
	fmt.Printf("✅ GitHub repository created: %s\n", repoInfo.HTMLURL)

	return nil
}

// initializeLocalRepository initializes local repository
func (ps *PushService) initializeLocalRepository(ctx context.Context, config *models.ProjectConfig) error {
	// Initialize Git repository
	if err := ps.gitService.InitializeRepository(ctx); err != nil {
		return err
	}

	// Add files
	if err := ps.gitService.AddAllFiles(ctx); err != nil {
		return err
	}

	// Create initial commit
	commitMessage := fmt.Sprintf("Initial commit for %s", config.Name)
	if config.HasTemplate() {
		commitMessage = fmt.Sprintf("Initial commit from template %s", config.Template.FullName)
	}

	if err := ps.gitService.CreateInitialCommit(ctx, commitMessage); err != nil {
		return err
	}

	return nil
}

// pushToRemoteRepository pushes to remote repository
func (ps *PushService) pushToRemoteRepository(ctx context.Context, repoInfo *github.RepositoryInfo) error {
	// Add remote repository
	if err := ps.gitService.AddRemote(ctx, "origin", repoInfo.CloneURL); err != nil {
		return err
	}

	// Get current branch
	branch, err := ps.gitService.GetCurrentBranch(ctx)
	if err != nil {
		branch = "main" // Default
	}

	// Execute push
	if err := ps.gitService.PushToRemote(ctx, "origin", branch); err != nil {
		return models.NewGitHubError(
			fmt.Sprintf("Failed to push to GitHub. Repository URL: %s", repoInfo.HTMLURL),
			err,
		)
	}

	return nil
}

// SetupGitConfiguration sets up Git configuration
func (ps *PushService) SetupGitConfiguration(ctx context.Context, name, email string) error {
	return ps.gitService.ConfigureUserInfo(ctx, name, email)
}
