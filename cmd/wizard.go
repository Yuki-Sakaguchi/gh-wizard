package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	
	// 2. Git初期化
	gitInit := exec.CommandContext(ctx, "git", "init")
	gitInit.Dir = config.LocalPath
	if err := gitInit.Run(); err != nil {
		return models.NewValidationError(fmt.Sprintf("Git初期化に失敗しました: %v", err))
	}
	
	// 3. テンプレートからファイルコピー（該当する場合）
	if config.Template != nil {
		fmt.Printf("📦 テンプレート '%s' を適用中...\n", config.Template.FullName)
		// TODO: 実際のテンプレートコピー処理を実装
		// 現時点では基本的なREADME.mdを作成
		if err := wr.createBasicFiles(config); err != nil {
			return err
		}
	} else {
		// テンプレートなしの場合は基本ファイルのみ作成
		if err := wr.createBasicFiles(config); err != nil {
			return err
		}
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
	args := []string{"repo", "create", config.Name}
	
	if config.Description != "" {
		args = append(args, "--description", config.Description)
	}
	
	if config.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}
	
	// ローカルリポジトリとして設定
	args = append(args, "--source", config.LocalPath)
	
	createCmd := exec.CommandContext(ctx, "gh", args...)
	createCmd.Dir = config.LocalPath
	
	if err := createCmd.Run(); err != nil {
		return models.NewGitHubError("GitHubリポジトリの作成に失敗しました", err)
	}
	
	return nil
}
