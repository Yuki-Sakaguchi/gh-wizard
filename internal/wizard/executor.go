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
)

// ProjectExecutor はプロジェクト作成を実行する
type ProjectExecutor struct {
	githubClient github.Client
}

// NewProjectExecutor は新しいExecutorを作成する
func NewProjectExecutor(githubClient github.Client) *ProjectExecutor {
	return &ProjectExecutor{
		githubClient: githubClient,
	}
}

// Execute はプロジェクト作成を実行する
func (pe *ProjectExecutor) Execute(ctx context.Context, config *models.ProjectConfig) error {
	// 1. ローカルディレクトリの作成
	fmt.Println("✓ テンプレートからディレクトリを作成中...")
	if err := pe.createLocalDirectory(ctx, config); err != nil {
		return fmt.Errorf("ローカルディレクトリの作成に失敗: %w", err)
	}

	// 2. GitHubリポジトリの作成（オプション）
	if config.CreateGitHub {
		fmt.Println("✓ GitHubリポジトリを作成中...")
		if err := pe.createGitHubRepository(ctx, config); err != nil {
			return fmt.Errorf("GitHubリポジトリの作成に失敗: %w", err)
		}

		fmt.Println("✓ ローカルリポジトリをプッシュ中...")
		if err := pe.pushToGitHub(ctx, config); err != nil {
			return fmt.Errorf("GitHubへのプッシュに失敗: %w", err)
		}
	}

	fmt.Println("✅ 完了！ プロジェクトの準備ができました")
	pe.printSuccess(config)

	return nil
}

// createLocalDirectory はテンプレートからローカルディレクトリを作成する
func (pe *ProjectExecutor) createLocalDirectory(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()

	// ディレクトリが既に存在するかチェック
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		return models.NewProjectError(
			fmt.Sprintf("ディレクトリ '%s' は既に存在します", targetPath),
			nil,
		)
	}

	if config.HasTemplate() {
		return pe.cloneFromTemplate(ctx, config)
	} else {
		return pe.createEmptyProject(ctx, config)
	}
}

// createEmptyProject は空のプロジェクトを作成する
func (pe *ProjectExecutor) createEmptyProject(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()

	// ディレクトリを作成
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return models.NewProjectError(
			"ディレクトリの作成に失敗しました",
			err,
		)
	}

	// README.md を作成
	readmePath := filepath.Join(targetPath, "README.md")
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", config.Name, config.Description)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return models.NewProjectError(
			"README.md の作成に失敗しました",
			err,
		)
	}

	// Gitリポジトリとして初期化
	return pe.initializeGitRepository(ctx, targetPath)
}

// printSuccess は成功メッセージを表示する
func (pe *ProjectExecutor) printSuccess(config *models.ProjectConfig) {
	fmt.Println()
	fmt.Println("🎉 プロジェクトが正常に作成されました！")
	fmt.Println()

	// プロジェクト情報の表示
	summary := config.GetDisplaySummary()
	for _, line := range summary {
		fmt.Printf("  %s\n", line)
	}

	fmt.Println()

	// 次のステップを提案
	fmt.Println("📝 次のステップ:")
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
			fmt.Println("  # プロジェクトのREADMEを確認してください")
		}
	}

	if config.CreateGitHub {
		repoURL := fmt.Sprintf("https://github.com/%s/%s", getCurrentUser(), config.Name)
		fmt.Printf("\n🔗 GitHubリポジトリ: %s\n", repoURL)
	}
}

// getCurrentUser は現在のGitHubユーザー名を取得する
func getCurrentUser() string {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "unknown" // フォールバック
	}
	return strings.TrimSpace(string(output))
}
