package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CheckGitInstalled Gitがインストールされているかチェック
func CheckGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("gitがインストールされていません: %w", err)
	}
	return nil
}

// GetGitVersion Gitのバージョンを取得
func GetGitVersion() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("gitのバージョン取得に失敗: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CheckGitVersion Gitのバージョンが要件を満たすかチェック
func CheckGitVersion() error {
	version, err := GetGitVersion()
	if err != nil {
		return err
	}

	// 最低限gitがインストールされていればOKとする
	if !strings.Contains(version, "git version") {
		return fmt.Errorf("無効なgitバージョン: %s", version)
	}

	return nil
}

// CheckGHInstalled は GitHub CLI がインストールされているかどうかをチェックする
func CheckGHInstalled() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("GitHub CLI がインストールされていません。https://cli.github.com/ からインストールしてください。")
	}
	return nil
}

// GitGHVersion は GitHub CLI のバージョンを取得する
func GitGHVersion() (string, error) {
	cmd := exec.Command("gh", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("GitHub CLI のバージョンの取得に失敗しました: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "不明", nil
}

// CheckGHVersion は GitHub CLI のバージョンが要件を満たすかどうかチェックする
func CheckGHVersion() error {
	version, err := GitGHVersion()
	if err != nil {
		return err
	}

	if !strings.Contains(version, "gh version") {
		return fmt.Errorf("GitHub CLI のバージョン形式が不正です: %s", version)
	}

	return nil
}

// GitService はGit操作を管理するサービス
type GitService struct {
	workingDir string
}

// NewGitService は新しいGitServiceを作成する
func NewGitService(workingDir string) *GitService {
	return &GitService{
		workingDir: workingDir,
	}
}

// CheckGitInstallation はGitがインストールされているかチェックする
func (gs *GitService) CheckGitInstallation() error {
	return CheckGitInstalled()
}

// InitializeRepository はGitリポジトリを初期化する
func (gs *GitService) InitializeRepository(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = gs.workingDir
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gitリポジトリの初期化に失敗: %w", err)
	}
	
	return nil
}

// AddAllFiles はすべてのファイルをステージングエリアに追加する
func (gs *GitService) AddAllFiles(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = gs.workingDir
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ファイルの追加に失敗: %w", err)
	}
	
	return nil
}

// CreateInitialCommit は初期コミットを作成する
func (gs *GitService) CreateInitialCommit(ctx context.Context, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = gs.workingDir
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("初期コミットの作成に失敗: %w", err)
	}
	
	return nil
}

// GetCurrentBranch は現在のブランチ名を取得する
func (gs *GitService) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = gs.workingDir
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("現在のブランチの取得に失敗: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// AddRemote はリモートリポジトリを追加する
func (gs *GitService) AddRemote(ctx context.Context, name, url string) error {
	cmd := exec.CommandContext(ctx, "git", "remote", "add", name, url)
	cmd.Dir = gs.workingDir
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("リモートリポジトリの追加に失敗: %w", err)
	}
	
	return nil
}

// PushToRemote はリモートリポジトリにプッシュする
func (gs *GitService) PushToRemote(ctx context.Context, remoteName, branchName string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "-u", remoteName, branchName)
	cmd.Dir = gs.workingDir
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("リモートリポジトリへのプッシュに失敗: %w", err)
	}
	
	return nil
}
