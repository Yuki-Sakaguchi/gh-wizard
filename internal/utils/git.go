package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// GitService はGit操作を提供する
type GitService struct {
	workingDir string
}

// NewGitService は新しいGitサービスを作成する
func NewGitService(workingDir string) *GitService {
	return &GitService{workingDir: workingDir}
}

// InitializeRepository はGitリポジトリを初期化する
func (gs *GitService) InitializeRepository(ctx context.Context) error {
	// git init
	if err := gs.runGitCommand(ctx, "init"); err != nil {
		return models.NewProjectError("Git初期化に失敗しました", err)
	}

	// デフォルトブランチをmainに設定
	if err := gs.runGitCommand(ctx, "branch", "-M", "main"); err != nil {
		// 古いGitの場合はスキップ
		fmt.Println("警告: デフォルトブランチの設定をスキップしました")
	}

	return nil
}

// AddAllFiles は全ファイルをステージングエリアに追加する
func (gs *GitService) AddAllFiles(ctx context.Context) error {
	return gs.runGitCommand(ctx, "add", ".")
}

// CreateInitialCommit は初期コミットを作成する
func (gs *GitService) CreateInitialCommit(ctx context.Context, message string) error {
	if message == "" {
		message = "Initial commit"
	}

	return gs.runGitCommand(ctx, "commit", "-m", message)
}

// AddRemote はリモートリポジトリを追加する
func (gs *GitService) AddRemote(ctx context.Context, name, url string) error {
	return gs.runGitCommand(ctx, "remote", "add", name, url)
}

// PushToRemote はリモートリポジトリにプッシュする
func (gs *GitService) PushToRemote(ctx context.Context, remote, branch string) error {
	return gs.runGitCommand(ctx, "push", "-u", remote, branch)
}

// SetUpstreamBranch はアップストリームブランチを設定する
func (gs *GitService) SetUpstreamBranch(ctx context.Context, remote, branch string) error {
	return gs.runGitCommand(ctx, "branch", "--set-upstream-to", fmt.Sprintf("%s/%s", remote, branch))
}

// GetCurrentBranch は現在のブランチ名を取得する
func (gs *GitService) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", models.NewProjectError("現在のブランチの取得に失敗しました", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// CheckGitInstallation はGitがインストールされているかチェックする
func (gs *GitService) CheckGitInstallation() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return models.NewProjectError(
			"Git がインストールされていません。https://git-scm.com/ からインストールしてください",
			err,
		)
	}
	return nil
}

// ConfigureUserInfo はユーザー情報を設定する（必要な場合）
func (gs *GitService) ConfigureUserInfo(ctx context.Context, name, email string) error {
	// グローバル設定をチェック
	if err := gs.checkGitConfig(ctx, "user.name"); err != nil {
		if name != "" {
			if err := gs.runGitCommand(ctx, "config", "user.name", name); err != nil {
				return models.NewProjectError("Gitユーザー名の設定に失敗しました", err)
			}
		}
	}

	if err := gs.checkGitConfig(ctx, "user.email"); err != nil {
		if email != "" {
			if err := gs.runGitCommand(ctx, "config", "user.email", email); err != nil {
				return models.NewProjectError("Gitメールアドレスの設定に失敗しました", err)
			}
		}
	}

	return nil
}

// runGitCommand はGitコマンドを実行する
func (gs *GitService) runGitCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = gs.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// checkGitConfig はGit設定をチェックする
func (gs *GitService) checkGitConfig(ctx context.Context, key string) error {
	cmd := exec.CommandContext(ctx, "git", "config", key)
	cmd.Dir = gs.workingDir

	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("設定 %s が見つかりません", key)
	}

	return nil
}
