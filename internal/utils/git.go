package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// GitService provides Git operations
type GitService struct {
	workingDir string
}

// NewGitService creates a new Git service
func NewGitService(workingDir string) *GitService {
	return &GitService{workingDir: workingDir}
}

// InitializeRepository initializes Git repository
func (gs *GitService) InitializeRepository(ctx context.Context) error {
	// git init
	if err := gs.runGitCommand(ctx, "init"); err != nil {
		return models.NewProjectError("Failed to initialize Git", err)
	}

	// Set default branch to main
	if err := gs.runGitCommand(ctx, "branch", "-M", "main"); err != nil {
		// Skip for older Git versions
		fmt.Println("Warning: Skipped default branch configuration")
	}

	return nil
}

// AddAllFiles adds all files to staging area
func (gs *GitService) AddAllFiles(ctx context.Context) error {
	return gs.runGitCommand(ctx, "add", ".")
}

// CreateInitialCommit creates initial commit
func (gs *GitService) CreateInitialCommit(ctx context.Context, message string) error {
	if message == "" {
		message = "Initial commit"
	}

	return gs.runGitCommand(ctx, "commit", "-m", message)
}

// AddRemote adds remote repository
func (gs *GitService) AddRemote(ctx context.Context, name, url string) error {
	return gs.runGitCommand(ctx, "remote", "add", name, url)
}

// PushToRemote pushes to remote repository
func (gs *GitService) PushToRemote(ctx context.Context, remote, branch string) error {
	return gs.runGitCommand(ctx, "push", "-u", remote, branch)
}

// SetUpstreamBranch sets upstream branch
func (gs *GitService) SetUpstreamBranch(ctx context.Context, remote, branch string) error {
	return gs.runGitCommand(ctx, "branch", "--set-upstream-to", fmt.Sprintf("%s/%s", remote, branch))
}

// GetCurrentBranch gets current branch name
func (gs *GitService) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", models.NewProjectError("Failed to get current branch", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckGitInstallation checks if Git is installed
func (gs *GitService) CheckGitInstallation() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return models.NewProjectError(
			"Git is not installed. Please install it from https://git-scm.com/",
			err,
		)
	}
	return nil
}

// ConfigureUserInfo configures user information (if needed)
func (gs *GitService) ConfigureUserInfo(ctx context.Context, name, email string) error {
	// Check global configuration
	if err := gs.checkGitConfig(ctx, "user.name"); err != nil {
		if name != "" {
			if err := gs.runGitCommand(ctx, "config", "user.name", name); err != nil {
				return models.NewProjectError("Failed to configure Git username", err)
			}
		}
	}

	if err := gs.checkGitConfig(ctx, "user.email"); err != nil {
		if email != "" {
			if err := gs.runGitCommand(ctx, "config", "user.email", email); err != nil {
				return models.NewProjectError("Failed to configure Git email address", err)
			}
		}
	}

	return nil
}

// runGitCommand executes Git command
func (gs *GitService) runGitCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = gs.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// checkGitConfig checks Git configuration
func (gs *GitService) checkGitConfig(ctx context.Context, key string) error {
	cmd := exec.CommandContext(ctx, "git", "config", key)
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("configuration %s not found", key)
	}

	return nil
}
