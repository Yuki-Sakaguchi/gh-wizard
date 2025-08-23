package wizard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/utils"
)

// ProjectExecutor executes project creation
type ProjectExecutor struct {
	githubClient github.Client
}

// NewProjectExecutor creates a new Executor
func NewProjectExecutor(githubClient github.Client) *ProjectExecutor {
	return &ProjectExecutor{
		githubClient: githubClient,
	}
}

// Execute executes project creation
func (pe *ProjectExecutor) Execute(ctx context.Context, config *models.ProjectConfig) error {
	// 1. Create local directory
	fmt.Println("✓ Creating directory from template...")
	if err := pe.createLocalDirectory(ctx, config); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// 2. Create GitHub repository (optional)
	if config.CreateGitHub {
		fmt.Println("✓ Creating GitHub repository...")
		if err := pe.createGitHubRepository(ctx, config); err != nil {
			return fmt.Errorf("failed to create GitHub repository: %w", err)
		}

		fmt.Println("✓ Pushing local repository...")
		if err := pe.pushToGitHub(ctx, config); err != nil {
			return fmt.Errorf("failed to push to GitHub: %w", err)
		}
	}

	fmt.Println("✅ Complete! Your project is ready")
	pe.printSuccess(config)

	return nil
}

// createLocalDirectory creates a local directory from template
func (pe *ProjectExecutor) createLocalDirectory(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()

	// Check if directory already exists
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return models.NewProjectError(
			fmt.Sprintf("directory '%s' already exists", targetPath),
			nil,
		)
	}

	if config.HasTemplate() {
		return pe.cloneFromTemplate(ctx, config)
	} else {
		return pe.createEmptyProject(ctx, config)
	}
}

// createEmptyProject creates an empty project
func (pe *ProjectExecutor) createEmptyProject(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()

	// Create directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return models.NewProjectError(
			"failed to create directory",
			err,
		)
	}

	// Create README.md
	readmePath := filepath.Join(targetPath, "README.md")
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", config.Name, config.Description)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return models.NewProjectError(
			"failed to create README.md",
			err,
		)
	}

	// Initialize as Git repository
	return pe.initializeGitRepository(ctx, targetPath)
}

// printSuccess displays success message
func (pe *ProjectExecutor) printSuccess(config *models.ProjectConfig) {
	fmt.Println()
	fmt.Println("🎉 Project created successfully!")
	fmt.Println()

	// Display project information
	summary := config.GetDisplaySummary()
	for _, line := range summary {
		fmt.Printf("  %s\n", line)
	}

	fmt.Println()

	// Suggest next steps
	fmt.Println("📝 Next steps:")
	fmt.Printf("  cd %s\n", config.Name)

	if config.Template != nil && config.Template.Language != "" {
		switch config.Template.Language {
		case "JavaScript", "TypeScript":
			fmt.Println("  npm install")
			fmt.Println("  npm run dev")
		case "Go":
			fmt.Println("  go mod tidy")
			fmt.Println("  go run main.go")
		case "Python":
			fmt.Println("  pip install -r requirements.txt")
			fmt.Println("  python main.py")
		default:
			fmt.Println("  # Please check the project README")
		}
	}

	if config.CreateGitHub {
		repoURL := fmt.Sprintf("https://github.com/%s/%s", getCurrentUser(), config.Name)
		fmt.Printf("\n🔗 GitHub repository: %s\n", repoURL)
	}
}

// getCurrentUser gets the current GitHub username
func getCurrentUser() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "unknown" // Fallback
	}
	return strings.TrimSpace(string(output))
}

// cloneFromTemplate creates a project from template
func (pe *ProjectExecutor) cloneFromTemplate(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()

	// Create from template using GitHub CLI
	repoURL := config.Template.CloneURL
	if repoURL == "" {
		repoURL = config.Template.GetRepoURL()
	}

	cmd := exec.CommandContext(ctx, "gh", "repo", "create", config.Name,
		"--template", config.Template.FullName,
		"--clone",
		"--local",
		"--destination", targetPath)

	if config.IsPrivate {
		cmd.Args = append(cmd.Args, "--private")
	} else {
		cmd.Args = append(cmd.Args, "--public")
	}

	if config.Description != "" {
		cmd.Args = append(cmd.Args, "--description", config.Description)
	}

	if err := cmd.Run(); err != nil {
		// If GitHub CLI fails, try regular git clone
		return pe.fallbackClone(ctx, config)
	}

	return nil
}

// fallbackClone is fallback when GitHub CLI fails
func (pe *ProjectExecutor) fallbackClone(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()
	repoURL := config.Template.CloneURL
	if repoURL == "" {
		repoURL = config.Template.GetRepoURL()
	}

	// For local directories, perform copy operation
	if isLocalPath(repoURL) {
		return pe.copyFromLocalTemplate(repoURL, targetPath)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, targetPath)
	if err := cmd.Run(); err != nil {
		return models.NewProjectError(
			fmt.Sprintf("failed to clone template: %s", repoURL),
			err,
		)
	}

	// Remove .git directory and initialize as new repository
	gitDir := filepath.Join(targetPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return models.NewProjectError("failed to remove existing .git directory", err)
	}

	// Initialize as new Git repository
	return pe.initializeGitRepository(ctx, targetPath)
}

// isLocalPath determines if the path is a local path
func isLocalPath(path string) bool {
	return !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") && !strings.HasPrefix(path, "git@")
}

// copyFromLocalTemplate copies files from local template
func (pe *ProjectExecutor) copyFromLocalTemplate(sourcePath, targetPath string) error {
	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return models.NewProjectError("failed to create target directory", err)
	}

	// Copy files from source directory to target directory
	if err := copyDir(sourcePath, targetPath); err != nil {
		return models.NewProjectError("failed to copy template files", err)
	}

	// Initialize as Git repository
	return pe.initializeGitRepository(context.Background(), targetPath)
}

// copyDir recursively copies directories
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory if it's a directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file if it's a file
			return copyFile(path, dstPath)
		}
	})
}

// copyFile copies a file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = srcFile.WriteTo(dstFile)
	return err
}

// initializeGitRepository initializes Git repository
func (pe *ProjectExecutor) initializeGitRepository(ctx context.Context, targetPath string) error {
	gitService := utils.NewGitService(targetPath)

	// Initialize Git
	if err := gitService.InitializeRepository(ctx); err != nil {
		return models.NewProjectError("failed to initialize Git repository", err)
	}

	// Add all files
	if err := gitService.AddAllFiles(ctx); err != nil {
		return models.NewProjectError("failed to add files", err)
	}

	// Create initial commit
	if err := gitService.CreateInitialCommit(ctx, "Initial commit"); err != nil {
		return models.NewProjectError("failed to create initial commit", err)
	}

	return nil
}

// createGitHubRepository creates a GitHub repository
func (pe *ProjectExecutor) createGitHubRepository(ctx context.Context, config *models.ProjectConfig) error {
	// Create repository using GitHub CLI
	cmd := exec.CommandContext(ctx, "gh", "repo", "create", config.Name)

	if config.IsPrivate {
		cmd.Args = append(cmd.Args, "--private")
	} else {
		cmd.Args = append(cmd.Args, "--public")
	}

	if config.Description != "" {
		cmd.Args = append(cmd.Args, "--description", config.Description)
	}

	if err := cmd.Run(); err != nil {
		return models.NewProjectError("failed to create GitHub repository", err)
	}

	return nil
}

// pushToGitHub pushes local repository to GitHub
func (pe *ProjectExecutor) pushToGitHub(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()
	gitService := utils.NewGitService(targetPath)

	// Add remote repository
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", getCurrentUser(), config.Name)
	if err := gitService.AddRemote(ctx, "origin", repoURL); err != nil {
		return models.NewProjectError("failed to add remote repository", err)
	}

	// Get current branch
	branch, err := gitService.GetCurrentBranch(ctx)
	if err != nil {
		// Use default main if branch cannot be retrieved
		branch = "main"
	}

	// Push
	if err := gitService.PushToRemote(ctx, "origin", branch); err != nil {
		return models.NewProjectError("failed to push to GitHub", err)
	}

	return nil
}
