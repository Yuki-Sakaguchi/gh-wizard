package utils

import (
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
