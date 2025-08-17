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

// cloneFromTemplate はテンプレートからプロジェクトを作成する
func (pe *ProjectExecutor) cloneFromTemplate(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()
	
	// GitHub CLI を使ってテンプレートから作成
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
		// GitHub CLIが失敗した場合は、通常のgit cloneを試行
		return pe.fallbackClone(ctx, config)
	}
	
	return nil
}

// fallbackClone はGitHub CLIが失敗した場合のフォールバック
func (pe *ProjectExecutor) fallbackClone(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()
	repoURL := config.Template.CloneURL
	if repoURL == "" {
		repoURL = config.Template.GetRepoURL()
	}
	
	// ローカルディレクトリの場合はコピーを実行
	if isLocalPath(repoURL) {
		return pe.copyFromLocalTemplate(repoURL, targetPath)
	}
	
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, targetPath)
	if err := cmd.Run(); err != nil {
		return models.NewProjectError(
			fmt.Sprintf("テンプレートのクローンに失敗しました: %s", repoURL),
			err,
		)
	}
	
	// .git ディレクトリを削除して新しいリポジトリとして初期化
	gitDir := filepath.Join(targetPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return models.NewProjectError("既存の.gitディレクトリの削除に失敗", err)
	}
	
	// 新しいGitリポジトリとして初期化
	return pe.initializeGitRepository(ctx, targetPath)
}

// isLocalPath はパスがローカルパスかどうかを判定する
func isLocalPath(path string) bool {
	return !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") && !strings.HasPrefix(path, "git@")
}

// copyFromLocalTemplate はローカルテンプレートからファイルをコピーする
func (pe *ProjectExecutor) copyFromLocalTemplate(sourcePath, targetPath string) error {
	// ターゲットディレクトリを作成
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return models.NewProjectError("ターゲットディレクトリの作成に失敗", err)
	}
	
	// ソースディレクトリからターゲットディレクトリにファイルをコピー
	if err := copyDir(sourcePath, targetPath); err != nil {
		return models.NewProjectError("テンプレートファイルのコピーに失敗", err)
	}
	
	// Gitリポジトリとして初期化
	return pe.initializeGitRepository(context.Background(), targetPath)
}

// copyDir はディレクトリを再帰的にコピーする
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 相対パスを計算
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			// ディレクトリの場合は作成
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// ファイルの場合はコピー
			return copyFile(path, dstPath)
		}
	})
}

// copyFile はファイルをコピーする
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	// ディレクトリが存在しない場合は作成
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

// initializeGitRepository はGitリポジトリを初期化する
func (pe *ProjectExecutor) initializeGitRepository(ctx context.Context, targetPath string) error {
	gitService := utils.NewGitService(targetPath)
	
	// Git初期化
	if err := gitService.InitializeRepository(ctx); err != nil {
		return models.NewProjectError("Gitリポジトリの初期化に失敗", err)
	}
	
	// 全ファイルを追加
	if err := gitService.AddAllFiles(ctx); err != nil {
		return models.NewProjectError("ファイルの追加に失敗", err)
	}
	
	// 初期コミット
	if err := gitService.CreateInitialCommit(ctx, "Initial commit"); err != nil {
		return models.NewProjectError("初期コミットの作成に失敗", err)
	}
	
	return nil
}

// createGitHubRepository はGitHubリポジトリを作成する
func (pe *ProjectExecutor) createGitHubRepository(ctx context.Context, config *models.ProjectConfig) error {
	// GitHub CLIを使ってリポジトリを作成
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
		return models.NewProjectError("GitHubリポジトリの作成に失敗", err)
	}
	
	return nil
}

// pushToGitHub はローカルリポジトリをGitHubにプッシュする
func (pe *ProjectExecutor) pushToGitHub(ctx context.Context, config *models.ProjectConfig) error {
	targetPath := config.GetLocalCreatePath()
	gitService := utils.NewGitService(targetPath)
	
	// リモートリポジトリを追加
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", getCurrentUser(), config.Name)
	if err := gitService.AddRemote(ctx, "origin", repoURL); err != nil {
		return models.NewProjectError("リモートリポジトリの追加に失敗", err)
	}
	
	// 現在のブランチを取得
	branch, err := gitService.GetCurrentBranch(ctx)
	if err != nil {
		// ブランチが取得できない場合はデフォルトのmainを使用
		branch = "main"
	}
	
	// プッシュ
	if err := gitService.PushToRemote(ctx, "origin", branch); err != nil {
		return models.NewProjectError("GitHubへのプッシュに失敗", err)
	}
	
	return nil
}
