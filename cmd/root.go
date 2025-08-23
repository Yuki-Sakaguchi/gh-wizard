package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/wizard"
	"github.com/spf13/cobra"
)

var (
	templateFlag  string
	nameFlag      string
	dryRunFlag    bool
	yesFlag       bool
	classicUIFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "wizard",
	Short: "🔮 GitHub Repository Wizard",
	Long:  "Magically simple and intuitive GitHub repository creation wizard",
	RunE:  runWizard,
}

func init() {
	// Flag definitions
	rootCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Template to use (e.g. user/repo or 'none')")
	rootCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Project name (for non-interactive mode)")
	rootCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show configuration only without actual creation")
	rootCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip all confirmations")
	rootCmd.Flags().BoolVar(&classicUIFlag, "classic-ui", false, "Use classic multi-question UI instead of create-next-app style")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runWizard(cmd *cobra.Command, args []string) error {
	// Setup signal handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Channel to capture Ctrl+C (SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine for graceful shutdown
	go func() {
		<-sigChan
		fmt.Println("\n\n👋 Exiting...")
		cancel()
		os.Exit(0)
	}()

	runner := NewWizardRunner()

	// Check prerequisites
	if !dryRunFlag {
		if err := runner.checkPrerequisites(ctx); err != nil {
			return runner.handleError(err)
		}
	}

	// Fetch user's template repositories
	fmt.Println("🔍 Fetching your template repositories...")
	templates, templateErr := runner.githubClient.SearchPopularTemplates(ctx)
	if templateErr != nil {
		// Continue without templates if fetching fails
		fmt.Printf("⚠️  Failed to fetch templates: %v\n", templateErr)
		fmt.Println("Continuing without templates.")
		templates = []models.Template{}
	} else if len(templates) == 0 {
		fmt.Println("📭 No template repositories found")
		fmt.Println("💡 Set repositories as 'Template repository' on GitHub to display them here.")
	} else {
		fmt.Printf("✅ Found %d template repositories\n", len(templates))
	}

	var config *models.ProjectConfig
	var err error

	// Non-interactive mode or interactive mode
	if nameFlag != "" || templateFlag != "" {
		// Non-interactive mode
		config, err = runner.runNonInteractiveMode(templates, templateFlag, nameFlag)
	} else {
		// Interactive mode
		config, err = runner.runInteractiveMode(templates)
	}

	if err != nil {
		return runner.handleError(err)
	}

	// Display configuration
	runner.printConfiguration(config)

	if dryRunFlag {
		fmt.Println("🔍 Dry run mode: No actual creation will be performed")
		return nil
	}

	// Confirmation
	if !yesFlag {
		confirmed, err := runner.confirmConfiguration()
		if err != nil {
			return runner.handleError(err)
		}
		if !confirmed {
			fmt.Println("👋 Exiting...")
			return nil
		}
	}

	// Execute project creation
	if err := runner.createProject(ctx, config); err != nil {
		return runner.handleError(err)
	}

	fmt.Println("✨ Project successfully created!")
	return nil
}

// WizardRunner manages wizard execution
type WizardRunner struct {
	githubClient github.Client
}

// NewWizardRunner creates a new WizardRunner
func NewWizardRunner() *WizardRunner {
	return &WizardRunner{
		githubClient: github.NewClient(),
	}
}

// checkPrerequisites checks if required commands are available
func (wr *WizardRunner) checkPrerequisites(ctx context.Context) error {
	// Check git command availability
	if _, err := exec.LookPath("git"); err != nil {
		return models.NewValidationError("Git command not found. Please install Git.")
	}

	// Check GitHub CLI availability
	if _, err := exec.LookPath("gh"); err != nil {
		return models.NewValidationError("GitHub CLI (gh) not found. Please install from https://cli.github.com/.")
	}

	// Check GitHub CLI authentication status
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return models.NewValidationError("Not logged in to GitHub CLI. Please run 'gh auth login'.")
	}

	return nil
}

// runNonInteractiveMode runs in non-interactive mode
func (wr *WizardRunner) runNonInteractiveMode(templates []models.Template, templateFlag, nameFlag string) (*models.ProjectConfig, error) {
	if nameFlag == "" {
		return nil, models.NewValidationError("--name flag is required")
	}

	config := &models.ProjectConfig{
		Name:      nameFlag,
		LocalPath: "./" + nameFlag,
	}

	// Set template if specified
	if templateFlag != "" && templateFlag != "none" {
		for _, tmpl := range templates {
			if tmpl.FullName == templateFlag || tmpl.Name == templateFlag {
				config.Template = &tmpl
				break
			}
		}
		if config.Template == nil {
			return nil, models.NewValidationError(fmt.Sprintf("Specified template '%s' not found", templateFlag))
		}
	}

	return config, nil
}

// runInteractiveMode runs in interactive mode
func (wr *WizardRunner) runInteractiveMode(templates []models.Template) (*models.ProjectConfig, error) {
	// Use QuestionFlow from wizard package
	flow := wizard.NewQuestionFlow(templates)

	// Execute interactive questions with appropriate UI style
	var config *models.ProjectConfig
	var err error

	if classicUIFlag {
		// Use classic multi-question UI
		config, err = flow.Execute()
	} else {
		// Use create-next-app style UI (default)
		config, err = flow.ExecuteCreateNextAppStyle()
	}

	if err != nil {
		return nil, models.NewValidationError(fmt.Sprintf("Failed to execute questions: %v", err))
	}

	// Set LocalPath
	config.LocalPath = "./" + config.Name

	return config, nil
}

// handleError formats and displays errors appropriately
func (wr *WizardRunner) handleError(err error) error {
	// Special handling for Context cancellation (Ctrl+C)
	if err == context.Canceled {
		fmt.Println("\n👋 Exiting...")
		return nil // Don't treat as error
	}

	// Special handling for Survey interrupt error (Ctrl+C during questions)
	if strings.Contains(err.Error(), "interrupt") {
		fmt.Println("\n👋 Exiting...")
		return nil // Don't treat as error
	}

	if wizardErr, ok := err.(*models.WizardError); ok {
		if wizardErr.IsRetryable() {
			fmt.Fprintf(os.Stderr, "❌ Error: %s\n💡 Please wait and try again later\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())
		}
		return err
	}

	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", err.Error())
	return err
}

// printConfiguration displays configuration details
func (wr *WizardRunner) printConfiguration(config *models.ProjectConfig) {
	fmt.Println("📝 Configuration Review")
	fmt.Printf("✓ Project Name: %s\n", config.Name)

	if config.Description != "" {
		fmt.Printf("✓ Description:  %s\n", config.Description)
	}

	if config.Template != nil {
		fmt.Printf("✓ Template:     %s (%d⭐)\n", config.Template.FullName, config.Template.Stars)
	} else {
		fmt.Println("✓ Template:     None")
	}

	fmt.Printf("✓ Local Path:   %s\n", config.LocalPath)

	if config.CreateGitHub {
		if config.IsPrivate {
			fmt.Println("✓ Private:      True")
		} else {
			fmt.Println("✓ Private:      False")
		}
	}
}

// confirmConfiguration asks user to confirm configuration
func (wr *WizardRunner) confirmConfiguration() (bool, error) {
	// Add a blank line before the confirmation question
	fmt.Println()

	confirm := false
	prompt := &survey.Confirm{
		Message: "Create project with this configuration?",
		Default: false,
	}

	err := survey.AskOne(prompt, &confirm)
	return confirm, err
}

// createProject executes the actual project creation
func (wr *WizardRunner) createProject(ctx context.Context, config *models.ProjectConfig) error {
	fmt.Printf("🚀 Creating project '%s'...\n", config.Name)

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 1. Create local directory
	if err := os.MkdirAll(config.LocalPath, 0755); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to create directory: %v", err))
	}

	// 2. Copy files from template (if applicable)
	if config.Template != nil {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("📦 Applying template '%s'...\n", config.Template.FullName)
		if err := wr.copyTemplateFiles(ctx, config); err != nil {
			return err
		}
	} else {
		// Create basic files only if no template
		if err := wr.createBasicFiles(config); err != nil {
			return err
		}
	}

	// 3. Git initialization (completely remove template's .git first)
	gitDirPath := filepath.Join(config.LocalPath, ".git")
	if err := os.RemoveAll(gitDirPath); err != nil {
		fmt.Printf("⚠️  Failed to remove existing .git directory: %v\n", err)
	}

	gitInit := exec.CommandContext(ctx, "git", "init")
	gitInit.Dir = config.LocalPath
	if err := gitInit.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to initialize Git: %v", err))
	}

	// 4. Create GitHub repository (if applicable)
	if config.CreateGitHub {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fmt.Printf("🐙 Creating GitHub repository...\n")
		if err := wr.createGitHubRepository(ctx, config); err != nil {
			return err
		}
	}

	return nil
}

// copyTemplateFiles copies files from template repository
func (wr *WizardRunner) copyTemplateFiles(ctx context.Context, config *models.ProjectConfig) error {
	// Create temporary directory and clone template repository
	tempDir, err := os.MkdirTemp("", "gh-wizard-template-*")
	if err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to create temporary directory: %v", err))
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// Clone template repository
	cloneCmd := exec.CommandContext(ctx, "gh", "repo", "clone", config.Template.FullName, tempDir)
	if err := cloneCmd.Run(); err != nil {
		return models.NewGitHubError(fmt.Sprintf("Failed to clone template repository: %v", err), err)
	}

	// Copy files excluding .git directory
	if err := wr.copyDirectoryContents(tempDir, config.LocalPath, []string{".git"}); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to copy template files: %v", err))
	}

	// Update project name and description (if README.md exists)
	if err := wr.updateTemplateVariables(config); err != nil {
		// Continue template application even if error occurs
		fmt.Printf("⚠️  Failed to update template variables: %v\n", err)
	}

	return nil
}

// copyDirectoryContents copies directory contents to another directory (with exclusion list support)
func (wr *WizardRunner) copyDirectoryContents(srcDir, dstDir string, excludeDirs []string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Check for excluded directories
		for _, excludeDir := range excludeDirs {
			if strings.HasPrefix(relPath, excludeDir) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// Copy file
			return wr.copyFile(srcPath, dstPath)
		}
	})
}

// copyFile copies a file
func (wr *WizardRunner) copyFile(srcPath, dstPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// updateTemplateVariables updates variables within template
func (wr *WizardRunner) updateTemplateVariables(config *models.ProjectConfig) error {
	readmePath := filepath.Join(config.LocalPath, "README.md")

	// Check if README.md exists
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		// Create basic README if README.md doesn't exist
		return wr.createBasicFiles(config)
	}

	// Read README.md
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Replace template variables (simple example)
	contentStr := string(content)

	// Replace common template variables
	replacements := map[string]string{
		"{{PROJECT_NAME}}": config.Name,
		"{{project_name}}": config.Name,
		"{{DESCRIPTION}}":  config.Description,
		"{{description}}":  config.Description,
		"${PROJECT_NAME}":  config.Name,
		"${project_name}":  config.Name,
		"${DESCRIPTION}":   config.Description,
		"${description}":   config.Description,
	}

	for placeholder, value := range replacements {
		if value != "" { // Don't replace if value is empty
			contentStr = strings.ReplaceAll(contentStr, placeholder, value)
		}
	}

	// Write back updated content
	return os.WriteFile(readmePath, []byte(contentStr), 0644)
}

// createBasicFiles creates basic files
func (wr *WizardRunner) createBasicFiles(config *models.ProjectConfig) error {
	// Create README.md
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", config.Name, config.Description)
	readmePath := fmt.Sprintf("%s/README.md", config.LocalPath)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to create README.md: %v", err))
	}

	return nil
}

// createGitHubRepository creates a GitHub repository
func (wr *WizardRunner) createGitHubRepository(ctx context.Context, config *models.ProjectConfig) error {
	// Initial commit and file addition
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = config.LocalPath
	if err := addCmd.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to stage files: %v", err))
	}

	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", "Initial commit")
	commitCmd.Dir = config.LocalPath
	if err := commitCmd.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to create initial commit: %v", err))
	}

	// 1. Create GitHub repository (without push)
	args := []string{"repo", "create", config.Name}

	if config.Description != "" {
		args = append(args, "--description", config.Description)
	}

	if config.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	createCmd := exec.CommandContext(ctx, "gh", args...)
	createCmd.Dir = config.LocalPath

	if output, err := createCmd.CombinedOutput(); err != nil {
		fmt.Printf("Error output: %s\n", string(output))
		return models.NewGitHubError(fmt.Sprintf("Failed to create GitHub repository: %s", string(output)), err)
	}

	// 2. Get GitHub username
	userCmd := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login")
	userOutput, err := userCmd.Output()
	if err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to get GitHub username: %v", err))
	}
	username := strings.TrimSpace(string(userOutput))

	// 3. Add remote repository
	remoteCmd := exec.CommandContext(ctx, "git", "remote", "add", "origin", fmt.Sprintf("https://github.com/%s/%s.git", username, config.Name))
	remoteCmd.Dir = config.LocalPath
	if err := remoteCmd.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to add remote repository: %v", err))
	}

	// 4. Get current branch name
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchCmd.Dir = config.LocalPath
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return models.NewValidationError(fmt.Sprintf("Failed to get current branch name: %v", err))
	}
	currentBranch := strings.TrimSpace(string(branchOutput))

	// 5. Push current branch
	pushCmd := exec.CommandContext(ctx, "git", "push", "-u", "origin", currentBranch)
	pushCmd.Dir = config.LocalPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		fmt.Printf("Push error (%s): %s\n", currentBranch, string(output))
		return models.NewGitHubError(fmt.Sprintf("Failed to push to repository: %s", string(output)), err)
	}

	return nil
}
