package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
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

	runner := &WizardRunner{}

	// 前提条件チェック
	if !dryRunFlag {
		if err := runner.checkPrerequisites(ctx); err != nil {
			return runner.handleError(err)
		}
	}

	// テンプレート一覧を取得（モック実装）
	templates := []models.Template{
		{Name: "basic", FullName: "github/basic", Stars: 100},
		{Name: "react", FullName: "facebook/react-template", Stars: 500},
		{Name: "go", FullName: "golang/go-template", Stars: 200},
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
type WizardRunner struct{}

// checkPrerequisites は必要なコマンドが利用可能かチェック
func (wr *WizardRunner) checkPrerequisites(ctx context.Context) error {
	// TODO: 実際のコマンド存在チェックを実装
	// exec.LookPath("git") および exec.LookPath("gh") をチェック
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
	// TODO: 実際の対話式インターフェースを実装
	// survey パッケージを使用した対話的な入力処理
	return &models.ProjectConfig{
		Name:      "interactive-project",
		LocalPath: "./interactive-project",
	}, nil
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
	// TODO: 実際の確認処理を実装
	// survey.Confirm を使用
	return true, nil
}

// createProject は実際のプロジェクト作成を実行
func (wr *WizardRunner) createProject(ctx context.Context, config *models.ProjectConfig) error {
	// TODO: 実際のプロジェクト作成ロジックを実装
	// 1. ローカルディレクトリ作成
	// 2. Git初期化
	// 3. テンプレートからファイルコピー（該当する場合）
	// 4. GitHubリポジトリ作成（該当する場合）
	// 5. リモート追加とpush
	return nil
}
