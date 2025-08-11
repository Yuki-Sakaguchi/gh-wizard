package github

import (
	"context"
	"fmt"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// CreateRepositoryWithProgress はプログレス付きでリポジトリを作成する
func (c *client) CreateRepositoryWithProgress(ctx context.Context, state *models.WizardState, progressChan chan<- models.ExecutionMessage) error {
	plan := models.NewExecutionPlan(state)
	plan.StartTime = time.Now()
	
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
			Message:  "設定の検証に失敗しました",
		}
		return err
	}
	
	time.Sleep(1 * time.Second) // 検証処理をシミュレート
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
		return err
	}

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

	return nil
}

// validateConfiguration は設定を検証する
func (c *client) validateConfiguration(state *models.WizardState) error {
	if state.RepoConfig == nil {
		return fmt.Errorf("リポジトリ設定が未設定です")
	}
	
	if err := state.RepoConfig.Validate(); err != nil {
		return err
	}
	
	return nil
}

// createRepository は実際のリポジトリ作成を行う
func (c *client) createRepository(ctx context.Context, state *models.WizardState) (string, error) {
	// gh コマンドを構築
	args := state.RepoConfig.GetGHCommand(state.SelectedTemplate)
	
	// 実行時間をシミュレート
	time.Sleep(3 * time.Second)
	
	// 実際の実装では以下のようにコマンドを実行
	// cmd := exec.CommandContext(ctx, "gh", args...)
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	//     return "", fmt.Errorf("リポジトリ作成に失敗しました: %v, output: %s", err, string(output))
	// }
	
	// 現在はシミュレーション用のargs参照のみ（実際の実装では削除）
	_ = args
	
	// 成功をシミュレート
	repoURL := fmt.Sprintf("https://github.com/%s/%s", "your-username", state.RepoConfig.Name)
	return repoURL, nil
}

// setupTemplate はテンプレートを設定する
func (c *client) setupTemplate(ctx context.Context, state *models.WizardState) error {
	// テンプレート設定処理をシミュレート
	time.Sleep(2 * time.Second)
	return nil
}

// createReadme はREADME.mdを作成する
func (c *client) createReadme(ctx context.Context, state *models.WizardState) error {
	// README作成処理をシミュレート
	time.Sleep(1 * time.Second)
	return nil
}

// cloneRepository はリポジトリをクローンする
func (c *client) cloneRepository(ctx context.Context, state *models.WizardState, repoURL string) error {
	// クローン処理をシミュレート
	time.Sleep(3 * time.Second)
	
	// 実際の実装では以下のようにクローンを実行
	// cmd := exec.CommandContext(ctx, "git", "clone", repoURL)
	// return cmd.Run()
	
	return nil
}