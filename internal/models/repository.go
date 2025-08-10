package models

import "errors"

// Repository は GitHub のリポジトリ作成時の設定を表す
type RepositoryConfig struct {
	Name        string
	Description string
	IsPrivate   bool
	SholdClone  bool
	AddReadme   bool
}

// Validate は設定値の妥当性をチェックする
func (rc RepositoryConfig) Validate() error {
	if rc.Name == "" {
		return errors.New("リポジトリ名は必須です")
	}
	if len(rc.Name) > 100 {
		return errors.New("リポジトリ名は100文字以内にしてください")
	}
	return nil
}

// GetGHCommand は gh repo create のコマンドを生成する
func (rc RepositoryConfig) GetGHCommand(template *Template) []string {
	args := []string{"repo", "create", rc.Name}

	if template != nil {
		args = append(args, "--template", template.FullName)
	}

	if rc.Description != "" {
		args = append(args, "--description", rc.Description)
	}

	if rc.IsPrivate {
		args = append(args, "--private")
	} else {
		args = append(args, "--public")
	}

	if rc.SholdClone {
		args = append(args, "--clone")
	}

	if rc.AddReadme && template == nil {
		args = append(args, "--add-readme")
	}

	return args
}
