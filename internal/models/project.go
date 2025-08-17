package models

import "fmt"

type ProjectConfig struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Template     *Template `json:"template,omitempty`
	CreateGitHub bool      `json:"create_github"`
	IsPrivate    bool      `json:"is_private"`
	LocalPath    string    `json:"local_path`
}

// Validate は設定値の妥当性をチェックする
func (pc *ProjectConfig) Validate() error {
	if pc.Name == "" {
		return fmt.Errorf("プロジェクト名は必須です")
	}

	if len(pc.Name) > 100 {
		return fmt.Errorf("プロジェクト名は最大100文字までです")
	}

	if len(pc.Description) > 500 {
		return fmt.Errorf("説明は最大500文字までです")
	}

	return nil
}

// GetGitHubCreateCommand は gh repo create コマンドの引数を生成する
func (pc *ProjectConfig) GetGitHubCreateCommand() []string {
	args := []string{"repo", "create", pc.Name}

	if pc.Template != nil {
		args = append(args, "--template", pc.Template.FullName)
	}

	if pc.Description != "" {
		args = append(args, "--description", pc.Description)
	}

	if pc.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	args = append(args, "--clone")

	return args
}

// GetLocalCreatePath はローカルの作成パスを返す
func (pc *ProjectConfig) GetLocalCreatePath() string {
	if pc.LocalPath != "" {
		return pc.LocalPath
	}
	return "./" + pc.Name
}

// HasTemplate はテンプレートが設定されているかを返す
func (pc *ProjectConfig) HasTemplate() bool {
	return pc.Template != nil
}

// GetDisplaySummary は設定内容の表示用サマリーを返す
func (pc *ProjectConfig) GetDisplaySummary() []string {
	summary := []string{
		fmt.Sprintf("📦 プロジェクト名: %s", pc.Name),
	}

	if pc.Description != "" {
		summary = append(summary, fmt.Sprintf("📄 説明: %s", pc.Description))
	}

	if pc.Template != nil {
		summary = append(summary, fmt.Sprintf("📚 テンプレート: %s", pc.Template.FullName))
	} else {
		summary = append(summary, "📚 テンプレート: なし")
	}

	if pc.CreateGitHub {
		visibility := "🌐 パブリック"
		if pc.IsPrivate {
			visibility = "🔒 プライベート"
		}
		summary = append(summary, fmt.Sprintf("🐙 GitHub: 作成する (%s)", visibility))
	} else {
		summary = append(summary, "🐙 GitHub: 作成しない")
	}

	summary = append(summary, fmt.Sprintf("📁 ローカルパス: %s", pc.GetLocalCreatePath()))

	return summary
}
