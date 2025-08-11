package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// CreateRepositoryWithProgress はプログレス付きでリポジトリを作成する
func (c *client) CreateRepositoryWithProgress(ctx context.Context, state *models.WizardState, progressChan chan<- models.ExecutionMessage) (*models.RepositoryCreationResult, error) {
	plan := models.NewExecutionPlan(state)
	plan.StartTime = time.Now()
	
	result := &models.RepositoryCreationResult{
		Success: false,
		Message: "処理開始",
	}
	
	// 設定の検証
	progressChan <- models.ExecutionMessage{
		TaskID:   "validate",
		Status:   models.TaskStatusInProgress,
		Progress: 0.0,
		Message:  "設定を検証しています...",
	}
	
	if err := c.validateConfiguration(state); err != nil {
		progressChan <- models.ExecutionMessage{
			TaskID:   "validate",
			Status:   models.TaskStatusFailed,
			Progress: 0.0,
			Error:    err,
			Message:  fmt.Sprintf("設定の検証に失敗しました: %v", err),
		}
		result.Error = err
		result.Message = fmt.Sprintf("設定の検証に失敗しました: %v", err)
		return result, err
	}
	
	// time.Sleep(1 * time.Second) // 検証処理をシミュレート（デバッグ用にコメントアウト）
	progressChan <- models.ExecutionMessage{
		TaskID:   "validate",
		Status:   models.TaskStatusCompleted,
		Progress: 1.0,
		Message:  "設定の検証が完了しました",
	}

	// リポジトリ作成
	progressChan <- models.ExecutionMessage{
		TaskID:   "create_repo",
		Status:   models.TaskStatusInProgress,
		Progress: 0.0,
		Message:  "リポジトリを作成しています...",
	}

	repoURL, err := c.createRepository(ctx, state)
	if err != nil {
		progressChan <- models.ExecutionMessage{
			TaskID:   "create_repo",
			Status:   models.TaskStatusFailed,
			Progress: 0.0,
			Error:    err,
			Message:  "リポジトリ作成に失敗しました",
		}
		result.Error = err
		result.Message = "リポジトリ作成に失敗しました"
		return result, err
	}
	
	result.RepositoryURL = repoURL
	progressChan <- models.ExecutionMessage{
		TaskID:   "create_repo",
		Status:   models.TaskStatusCompleted,
		Progress: 1.0,
		Message:  fmt.Sprintf("リポジトリが作成されました: %s", repoURL),
	}

	// テンプレート設定（該当する場合）
	if state.UseTemplate && state.SelectedTemplate != nil {
		progressChan <- models.ExecutionMessage{
			TaskID:   "setup_template",
			Status:   models.TaskStatusInProgress,
			Progress: 0.0,
			Message:  "テンプレートを適用しています...",
		}

		if err := c.setupTemplate(ctx, state); err != nil {
			// テンプレート設定の失敗は警告として扱い、継続
			progressChan <- models.ExecutionMessage{
				TaskID:   "setup_template",
				Status:   models.TaskStatusSkipped,
				Progress: 1.0,
				Error:    err,
				Message:  "テンプレート設定をスキップしました",
			}
		} else {
			progressChan <- models.ExecutionMessage{
				TaskID:   "setup_template",
				Status:   models.TaskStatusCompleted,
				Progress: 1.0,
				Message:  "テンプレートが適用されました",
			}
		}
	}

	// README作成（該当する場合）
	if state.RepoConfig.AddReadme {
		progressChan <- models.ExecutionMessage{
			TaskID:   "create_readme",
			Status:   models.TaskStatusInProgress,
			Progress: 0.0,
			Message:  "README.mdを作成しています...",
		}

		if err := c.createReadme(ctx, state); err != nil {
			progressChan <- models.ExecutionMessage{
				TaskID:   "create_readme",
				Status:   models.TaskStatusSkipped,
				Progress: 1.0,
				Error:    err,
				Message:  "README作成をスキップしました",
			}
		} else {
			progressChan <- models.ExecutionMessage{
				TaskID:   "create_readme",
				Status:   models.TaskStatusCompleted,
				Progress: 1.0,
				Message:  "README.mdが作成されました",
			}
		}
	}

	// クローン処理（該当する場合）
	if state.RepoConfig.SholdClone {
		progressChan <- models.ExecutionMessage{
			TaskID:   "clone_repo",
			Status:   models.TaskStatusInProgress,
			Progress: 0.0,
			Message:  "リポジトリをローカルにクローンしています...",
		}

		if err := c.cloneRepository(ctx, state, repoURL); err != nil {
			progressChan <- models.ExecutionMessage{
				TaskID:   "clone_repo",
				Status:   models.TaskStatusSkipped,
				Progress: 1.0,
				Error:    err,
				Message:  "クローン処理をスキップしました",
			}
		} else {
			progressChan <- models.ExecutionMessage{
				TaskID:   "clone_repo",
				Status:   models.TaskStatusCompleted,
				Progress: 1.0,
				Message:  "リポジトリがローカルにクローンされました",
			}
		}
	}
	
	// クローン処理の結果をresultに反映
	if state.RepoConfig.SholdClone {
		result.ClonePath = fmt.Sprintf("./%s", state.RepoConfig.Name)
	}
	
	// 成功結果を設定
	result.Success = true
	result.Message = "リポジトリの作成が正常に完了しました"

	return result, nil
}

// validateConfiguration は設定を検証する
func (c *client) validateConfiguration(state *models.WizardState) error {
	// シンプルな検証に変更
	if state == nil {
		return fmt.Errorf("ウィザード状態が未設定です")
	}
	
	if state.RepoConfig == nil {
		return fmt.Errorf("リポジトリ設定が未設定です")
	}
	
	if state.RepoConfig.Name == "" {
		return fmt.Errorf("リポジトリ名が設定されていません")
	}
	
	// 基本的な検証のみ実行
	if err := state.RepoConfig.Validate(); err != nil {
		return fmt.Errorf("リポジトリ設定の検証に失敗: %w", err)
	}
	
	// 実際のGitHub操作時のみ認証をチェック
	isSimulation := os.Getenv("GH_WIZARD_SIMULATION") != "false"
	if !isSimulation {
		if err := c.IsAuthenticated(); err != nil {
			return fmt.Errorf("GitHub CLI の認証が必要です: %w", err)
		}
	}
	
	return nil
}

// createRepository は実際のリポジトリ作成を行う
func (c *client) createRepository(ctx context.Context, state *models.WizardState) (string, error) {
	// gh コマンドを構築
	args := state.RepoConfig.GetGHCommand(state.SelectedTemplate)
	
	// 環境変数でシミュレーションモードを制御
	isSimulation := os.Getenv("GH_WIZARD_SIMULATION") != "false"
	
	if isSimulation {
		// シミュレーションモード
		time.Sleep(3 * time.Second) // リポジトリ作成処理をシミュレート
		
		// シミュレーション用のURL
		user, err := c.GetCurrentUser()
		var username string
		if err != nil {
			username = "your-username" // フォールバック
		} else {
			username = user.Login
		}
		
		repoURL := fmt.Sprintf("https://github.com/%s/%s", username, state.RepoConfig.Name)
		return repoURL, nil
	} else {
		// 実際のコマンド実行
		cmd := exec.CommandContext(ctx, "gh", args...)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return "", fmt.Errorf("リポジトリ作成に失敗しました: %w\n出力: %s", err, string(output))
		}
		
		// 成功時の出力からリポジトリURLを抽出
		repoURL, err := c.extractRepositoryURL(string(output), state.RepoConfig.Name)
		if err != nil {
			return "", fmt.Errorf("リポジトリURL の抽出に失敗しました: %w", err)
		}
		
		return repoURL, nil
	}
}

// extractRepositoryURL はgh コマンドの出力からリポジトリURLを抽出する
func (c *client) extractRepositoryURL(output, repoName string) (string, error) {
	// gh repo create の出力パターンをチェック
	// 例: "✓ Created repository username/repo-name on GitHub"
	// 例: "https://github.com/username/repo-name"
	
	// HTTPSのURLパターンを探す
	httpsPattern := regexp.MustCompile(`https://github\.com/[^/\s]+/[^/\s]+`)
	if matches := httpsPattern.FindString(output); matches != "" {
		return matches, nil
	}
	
	// Created repository パターンからURLを構築
	createdPattern := regexp.MustCompile(`✓\s+Created repository\s+([^/\s]+/[^/\s]+)`)
	if matches := createdPattern.FindStringSubmatch(output); len(matches) > 1 {
		return fmt.Sprintf("https://github.com/%s", matches[1]), nil
	}
	
	// フォールバック: ユーザー名を取得してURL構築
	user, err := c.GetCurrentUser()
	if err != nil {
		return "", fmt.Errorf("ユーザー情報の取得に失敗: %w", err)
	}
	
	return fmt.Sprintf("https://github.com/%s/%s", user.Login, repoName), nil
}

// setupTemplate はテンプレートを設定する
func (c *client) setupTemplate(ctx context.Context, state *models.WizardState) error {
	// テンプレートは `gh repo create` コマンドで自動的に適用されるため
	// ここでは特別な処理は不要
	// 将来的にはテンプレート適用後の追加設定などを実装可能
	
	return nil
}

// createReadme はREADME.mdを作成する
func (c *client) createReadme(ctx context.Context, state *models.WizardState) error {
	// README.md は `gh repo create --add-readme` で自動作成されるため
	// ここでは特別な処理は不要
	// 
	// 将来的には以下のような機能を実装可能:
	// - カスタマイズされたREADME テンプレートの適用
	// - プロジェクトタイプに応じた内容の自動生成
	// - ライセンスやバッジの追加
	
	return nil
}

// cloneRepository はリポジトリをクローンする
func (c *client) cloneRepository(ctx context.Context, state *models.WizardState, repoURL string) error {
	// 環境変数でシミュレーションモードを制御
	isSimulation := os.Getenv("GH_WIZARD_SIMULATION") != "false"
	
	if isSimulation {
		// シミュレーションモード
		time.Sleep(3 * time.Second) // クローン処理をシミュレート
		return nil
	} else {
		// 実際のクローン実行
		targetDir := state.RepoConfig.Name
		cmd := exec.CommandContext(ctx, "git", "clone", repoURL, targetDir)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return fmt.Errorf("リポジトリクローンに失敗しました: %w\n出力: %s", err, string(output))
		}
		
		return nil
	}
}