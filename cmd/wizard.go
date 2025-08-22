package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/wizard"
)

var (
	templateFlag string
	nameFlag     string
	dryRunFlag   bool
	yesFlag      bool
)

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "対話式リポジトリ作成ウィザードを開始",
	Long:  "🔮 GitHub Repository Wizard\n\n魔法のように簡単で直感的なGitHubリポジトリ作成ウィザード",
	RunE:  runWizard,
}

func init() {
	rootCmd.AddCommand(wizardCmd)
	
	// フラグの定義
	wizardCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "使用するテンプレート (例: user/repo または 'none')")
	wizardCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "プロジェクト名 (非対話モード用)")
	wizardCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "実際の作成は行わず、設定のみ表示")
	wizardCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "全ての確認をスキップ")
}

func runWizard(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	runner := NewWizardRunner()

	// 前提条件チェック
	if !dryRunFlag {
		if err := runner.checkPrerequisites(ctx); err != nil {
			return runner.handleError(err)
		}
	}

	// ユーザー自身のテンプレートリポジトリを取得
	fmt.Println("🔍 あなたのテンプレートリポジトリを取得中...")
	templates, templateErr := runner.githubClient.SearchPopularTemplates(ctx)
	if templateErr != nil {
		// テンプレート取得に失敗した場合はテンプレートなしで続行
		fmt.Printf("⚠️  テンプレート取得に失敗しました: %v\n", templateErr)
		fmt.Println("テンプレートなしで続行します。")
		templates = []models.Template{}
	} else if len(templates) == 0 {
		fmt.Println("📭 テンプレートリポジトリが見つかりませんでした")
		fmt.Println("💡 GitHubでリポジトリを「Template repository」として設定すると、ここに表示されます。")
	} else {
		fmt.Printf("✅ %d個のテンプレートリポジトリを見つけました\n", len(templates))
	}

	var config *models.ProjectConfig
	var err error

	// ノンインタラクティブモードまたは対話モード
	if nameFlag != "" || templateFlag != "" {
		// ノンインタラクティブモード
		config, err = runner.runNonInteractiveMode(templates, templateFlag, nameFlag)
	} else {
		// 対話モード
		config, err = runner.runInteractiveMode(templates)
	}

	if err != nil {
		return runner.handleError(err)
	}

	// 設定表示
	runner.printConfiguration(config)

	if dryRunFlag {
		fmt.Println("🔍 ドライランモード: 実際の作成は行いません")
		return nil
	}

	// 確認
	if !yesFlag {
		confirmed, err := runner.confirmConfiguration()
		if err != nil {
			return runner.handleError(err)
		}
		if !confirmed {
			fmt.Println("⏹️  キャンセルされました")
			return nil
		}
	}

	// プロジェクト作成実行
	if err := runner.createProject(ctx, config); err != nil {
		return runner.handleError(err)
	}

	fmt.Println("✨ プロジェクトが正常に作成されました！")
	return nil
}

// WizardRunner はウィザードの実行を管理する構造体
type WizardRunner struct {
	githubClient github.Client
}

// NewWizardRunner は新しいWizardRunnerを作成する
func NewWizardRunner() *WizardRunner {
	return &WizardRunner{
		githubClient: github.NewClient(),
	}
}

// checkPrerequisites は必要なコマンドが利用可能かチェック
func (wr *WizardRunner) checkPrerequisites(ctx context.Context) error {
	// gitコマンドの存在確認
	if _, err := exec.LookPath("git"); err != nil {
		return models.NewValidationError("gitコマンドが見つかりません。Gitをインストールしてください。")
	}
	
	// GitHub CLIの存在確認
	if _, err := exec.LookPath("gh"); err != nil {
		return models.NewValidationError("GitHub CLI (gh) が見つかりません。https://cli.github.com/ からインストールしてください。")
	}
	
	// GitHub CLIの認証状態確認
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return models.NewValidationError("GitHub CLIにログインしていません。'gh auth login' を実行してください。")
	}
	
	return nil
}

// runNonInteractiveMode は非対話モードでの実行
func (wr *WizardRunner) runNonInteractiveMode(templates []models.Template, templateFlag, nameFlag string) (*models.ProjectConfig, error) {
	if nameFlag == "" {
		return nil, models.NewValidationError("--name フラグが必要です")
	}

	config := &models.ProjectConfig{
		Name:      nameFlag,
		LocalPath: "./" + nameFlag,
	}

	// テンプレート指定があれば設定
	if templateFlag != "" && templateFlag != "none" {
		for _, tmpl := range templates {
			if tmpl.FullName == templateFlag || tmpl.Name == templateFlag {
				config.Template = &tmpl
				break
			}
		}
		if config.Template == nil {
			return nil, models.NewValidationError(fmt.Sprintf("指定されたテンプレート '%s' が見つかりません", templateFlag))
		}
	}

	return config, nil
}

// runInteractiveMode は対話モードでの実行
func (wr *WizardRunner) runInteractiveMode(templates []models.Template) (*models.ProjectConfig, error) {
	// wizard パッケージの QuestionFlow を使用
	flow := wizard.NewQuestionFlow(templates)
	
	// 対話的な質問を実行
	config, err := flow.Execute()
	if err != nil {
		return nil, models.NewValidationError(fmt.Sprintf("質問の実行に失敗しました: %v", err))
	}
	
	// LocalPathを設定
	config.LocalPath = "./" + config.Name
	
	return config, nil
}

// handleError はエラーを適切にフォーマットして表示
func (wr *WizardRunner) handleError(err error) error {
	if wizardErr, ok := err.(*models.WizardError); ok {
		if wizardErr.IsRetryable() {
			fmt.Fprintf(os.Stderr, "❌ エラー: %s\n💡 しばらく待ってから再実行してください\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "❌ エラー: %s\n", err.Error())
		}
		return err
	}
	
	fmt.Fprintf(os.Stderr, "❌ エラー: %s\n", err.Error())
	return err
}

// printConfiguration は設定内容を表示
func (wr *WizardRunner) printConfiguration(config *models.ProjectConfig) {
	fmt.Println("📋 設定内容確認")
	fmt.Printf("📝 プロジェクト名: %s\n", config.Name)
	
	if config.Description != "" {
		fmt.Printf("📖 説明: %s\n", config.Description)
	}
	
	if config.Template != nil {
		fmt.Printf("📦 テンプレート: %s (%d⭐)\n", config.Template.FullName, config.Template.Stars)
	} else {
		fmt.Println("📦 テンプレート: なし")
	}
	
	fmt.Printf("📁 ローカルパス: %s\n", config.LocalPath)
	
	if config.CreateGitHub {
		if config.IsPrivate {
			fmt.Println("👁️  可視性: プライベート")
		} else {
			fmt.Println("👁️  可視性: パブリック")
		}
	}
}

// confirmConfiguration はユーザーに設定確認を求める
func (wr *WizardRunner) confirmConfiguration() (bool, error) {
	confirm := false
	prompt := &survey.Confirm{
		Message: "この設定でプロジェクトを作成しますか？",
		Default: false,
	}
	
	err := survey.AskOne(prompt, &confirm)
	return confirm, err
}

// createProject は実際のプロジェクト作成を実行
func (wr *WizardRunner) createProject(ctx context.Context, config *models.ProjectConfig) error {
	fmt.Printf("🚀 プロジェクト '%s' を作成中...\n", config.Name)
	
	// 1. ローカルディレクトリ作成
	if err := os.MkdirAll(config.LocalPath, 0755); err != nil {
		return models.NewValidationError(fmt.Sprintf("ディレクトリの作成に失敗しました: %v", err))
	}
	
	// 2. テンプレートからファイルコピー（該当する場合）
	if config.Template != nil {
		fmt.Printf("📦 テンプレート '%s' を適用中...\n", config.Template.FullName)
		if err := wr.copyTemplateFiles(ctx, config); err != nil {
			return err
		}
	} else {
		// テンプレートなしの場合は基本ファイルのみ作成
		if err := wr.createBasicFiles(config); err != nil {
			return err
		}
	}
	
	// 3. Git初期化（テンプレートの.gitを完全に削除してから）
	gitDirPath := filepath.Join(config.LocalPath, ".git")
	if err := os.RemoveAll(gitDirPath); err != nil {
		fmt.Printf("⚠️  既存の.gitディレクトリの削除に失敗: %v\n", err)
	}
	
	gitInit := exec.CommandContext(ctx, "git", "init")
	gitInit.Dir = config.LocalPath
	if err := gitInit.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Git初期化に失敗しました: %v", err))
	}
	
	// 4. GitHubリポジトリ作成（該当する場合）
	if config.CreateGitHub {
		fmt.Printf("🐙 GitHubリポジトリを作成中...\n")
		if err := wr.createGitHubRepository(ctx, config); err != nil {
			return err
		}
	}
	
	return nil
}

// copyTemplateFiles はテンプレートリポジトリからファイルをコピーする
func (wr *WizardRunner) copyTemplateFiles(ctx context.Context, config *models.ProjectConfig) error {
	// 一時ディレクトリを作成してテンプレートリポジトリをクローン
	tempDir, err := os.MkdirTemp("", "gh-wizard-template-*")
	if err != nil {
		return models.NewValidationError(fmt.Sprintf("一時ディレクトリの作成に失敗しました: %v", err))
	}
	defer os.RemoveAll(tempDir) // クリーンアップ

	// テンプレートリポジトリをクローン
	cloneCmd := exec.CommandContext(ctx, "gh", "repo", "clone", config.Template.FullName, tempDir)
	if err := cloneCmd.Run(); err != nil {
		return models.NewGitHubError(fmt.Sprintf("テンプレートリポジトリのクローンに失敗しました: %v", err), err)
	}

	// .gitディレクトリを除外してファイルをコピー
	if err := wr.copyDirectoryContents(tempDir, config.LocalPath, []string{".git"}); err != nil {
		return models.NewValidationError(fmt.Sprintf("テンプレートファイルのコピーに失敗しました: %v", err))
	}

	// プロジェクト名とDescription を更新（README.mdが存在する場合）
	if err := wr.updateTemplateVariables(config); err != nil {
		// エラーが発生してもテンプレート適用は継続
		fmt.Printf("⚠️  テンプレート変数の更新に失敗しました: %v\n", err)
	}

	return nil
}

// copyDirectoryContents はディレクトリの内容を別のディレクトリにコピーする（除外リスト対応）
func (wr *WizardRunner) copyDirectoryContents(srcDir, dstDir string, excludeDirs []string) error {
	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 相対パスを取得
		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		// ルートディレクトリはスキップ
		if relPath == "." {
			return nil
		}

		// 除外ディレクトリのチェック
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
			// ディレクトリを作成
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			// ファイルをコピー
			return wr.copyFile(srcPath, dstPath)
		}
	})
}

// copyFile はファイルをコピーする
func (wr *WizardRunner) copyFile(srcPath, dstPath string) error {
	// ディレクトリが存在しない場合は作成
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

// updateTemplateVariables はテンプレート内の変数を更新する
func (wr *WizardRunner) updateTemplateVariables(config *models.ProjectConfig) error {
	readmePath := filepath.Join(config.LocalPath, "README.md")
	
	// README.mdが存在するかチェック
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		// README.mdが存在しない場合は基本的なREADMEを作成
		return wr.createBasicFiles(config)
	}

	// README.mdを読み取り
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// テンプレート変数を置換（簡単な例）
	contentStr := string(content)
	
	// 一般的なテンプレート変数を置換
	replacements := map[string]string{
		"{{PROJECT_NAME}}":    config.Name,
		"{{project_name}}":    config.Name,
		"{{DESCRIPTION}}":     config.Description,
		"{{description}}":     config.Description,
		"${PROJECT_NAME}":     config.Name,
		"${project_name}":     config.Name,
		"${DESCRIPTION}":      config.Description,
		"${description}":      config.Description,
	}

	for placeholder, value := range replacements {
		if value != "" { // 空の値の場合は置換しない
			contentStr = strings.ReplaceAll(contentStr, placeholder, value)
		}
	}

	// 更新された内容を書き戻し
	return os.WriteFile(readmePath, []byte(contentStr), 0644)
}

// createBasicFiles は基本ファイルを作成
func (wr *WizardRunner) createBasicFiles(config *models.ProjectConfig) error {
	// README.md作成
	readmeContent := fmt.Sprintf("# %s\n\n%s\n", config.Name, config.Description)
	readmePath := fmt.Sprintf("%s/README.md", config.LocalPath)
	
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return models.NewValidationError(fmt.Sprintf("README.mdの作成に失敗しました: %v", err))
	}
	
	return nil
}

// createGitHubRepository はGitHubリポジトリを作成
func (wr *WizardRunner) createGitHubRepository(ctx context.Context, config *models.ProjectConfig) error {
	// 初回コミットとファイル追加
	addCmd := exec.CommandContext(ctx, "git", "add", ".")
	addCmd.Dir = config.LocalPath
	if err := addCmd.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("ファイルのステージングに失敗しました: %v", err))
	}
	
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", "Initial commit")
	commitCmd.Dir = config.LocalPath
	if err := commitCmd.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("初回コミットに失敗しました: %v", err))
	}
	
	// GitHub リポジトリ作成 (--source ではなく現在のディレクトリから)
	args := []string{"repo", "create", config.Name}
	
	if config.Description != "" {
		args = append(args, "--description", config.Description)
	}
	
	if config.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}
	
	// --source を使わずに、現在のディレクトリから作成
	args = append(args, "--push")
	
	createCmd := exec.CommandContext(ctx, "gh", args...)
	createCmd.Dir = config.LocalPath
	
	if output, err := createCmd.CombinedOutput(); err != nil {
		fmt.Printf("エラー出力: %s\n", string(output))
		return models.NewGitHubError(fmt.Sprintf("GitHubリポジトリの作成に失敗しました: %s", string(output)), err)
	}
	
	return nil
}
