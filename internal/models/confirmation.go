package models

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ConfirmationAction は確認画面で実行可能なアクションを表す
type ConfirmationAction int

const (
	ActionModifySettings ConfirmationAction = iota
	ActionCancel
	ActionCreateRepository
)

// String はアクションの文字列表現を返す
func (a ConfirmationAction) String() string {
	switch a {
	case ActionModifySettings:
		return "設定修正"
	case ActionCancel:
		return "キャンセル"
	case ActionCreateRepository:
		return "リポジトリ作成"
	default:
		return "不明"
	}
}

// GetKey はアクションのショートカットキーを返す
func (a ConfirmationAction) GetKey() string {
	switch a {
	case ActionModifySettings:
		return "1"
	case ActionCancel:
		return "2"
	case ActionCreateRepository:
		return "3"
	default:
		return "?"
	}
}

// GetDescription はアクションの説明を返す
func (a ConfirmationAction) GetDescription() string {
	switch a {
	case ActionModifySettings:
		return "リポジトリ設定に戻って修正する"
	case ActionCancel:
		return "ウィザードを中断してメイン画面に戻る"
	case ActionCreateRepository:
		return "設定内容でリポジトリを作成する"
	default:
		return "不明なアクション"
	}
}

// ConfirmationItem は確認画面の個々の表示項目を表す
type ConfirmationItem struct {
	Label       string
	Value       string
	Description string
	Important   bool
	Warning     bool
}

// ConfirmationSection は確認画面のセクションを表す
type ConfirmationSection struct {
	Title      string
	Icon       string
	Items      []ConfirmationItem
	Warning    string
	HasWarning bool
}

// ConfirmationData は確認画面全体のデータを表す
type ConfirmationData struct {
	Sections      []ConfirmationSection
	Actions       []ConfirmationAction
	Warnings      []string
	RepositoryURL string
	EstimatedTime string
}

// BuildConfirmationData はウィザード状態から確認画面データを構築する（後方互換用）
func BuildConfirmationData(state *WizardState) *ConfirmationData {
	return BuildConfirmationDataWithClient(state, nil)
}

// BuildConfirmationDataWithClient は確認画面用のデータを構築する（GitHubクライアント付き）
func BuildConfirmationDataWithClient(state *WizardState, githubClient interface{}) *ConfirmationData {
	data := &ConfirmationData{
		Actions: []ConfirmationAction{
			ActionModifySettings,
			ActionCancel,
			ActionCreateRepository,
		},
	}

	// テンプレート情報セクション
	if state.UseTemplate && state.SelectedTemplate != nil {
		templateSection := buildTemplateSection(state.SelectedTemplate)
		data.Sections = append(data.Sections, templateSection)
	}

	// リポジトリ設定セクション
	if state.RepoConfig != nil {
		repoSection := buildRepositorySection(state.RepoConfig)
		data.Sections = append(data.Sections, repoSection)
	}

	// 作成先情報セクション
	destinationSection := buildDestinationSection(state, githubClient)
	data.Sections = append(data.Sections, destinationSection)

	// 警告メッセージを生成
	data.Warnings = buildWarnings(state)

	// リポジトリURLとその他の情報
	if state.RepoConfig != nil {
		data.RepositoryURL = fmt.Sprintf("https://github.com/%s/%s",
			getCurrentUserWithClient(githubClient), state.RepoConfig.Name)
		data.EstimatedTime = "約30秒"
	}

	return data
}

// buildTemplateSection はテンプレート情報セクションを構築する
func buildTemplateSection(template *Template) ConfirmationSection {
	items := []ConfirmationItem{
		{
			Label:     "名前",
			Value:     template.Name,
			Important: true,
		},
		{
			Label: "作成者",
			Value: template.Owner,
		},
		{
			Label: "言語",
			Value: template.Language,
		},
		{
			Label: "スター数",
			Value: fmt.Sprintf("⭐ %d", template.Stars),
		},
	}

	if template.Description != "" {
		items = append(items, ConfirmationItem{
			Label: "説明",
			Value: template.Description,
		})
	}

	if len(template.Topics) > 0 {
		items = append(items, ConfirmationItem{
			Label: "タグ",
			Value: strings.Join(template.Topics, ", "),
		})
	}

	if !template.UpdatedAt.IsZero() {
		items = append(items, ConfirmationItem{
			Label: "最終更新",
			Value: template.UpdatedAt.Format("2006-01-02"),
		})
	}

	return ConfirmationSection{
		Title: "使用テンプレート",
		Icon:  "📚",
		Items: items,
	}
}

// buildRepositorySection はリポジトリ設定セクションを構築する
func buildRepositorySection(config *RepositoryConfig) ConfirmationSection {
	items := []ConfirmationItem{
		{
			Label:     "リポジトリ名",
			Value:     config.Name,
			Important: true,
		},
		{
			Label: "公開設定",
			Value: func() string {
				if config.IsPrivate {
					return "🔒 プライベート（非公開）"
				}
				return "🌐 パブリック（公開）"
			}(),
			Important: true,
		},
	}

	if config.Description != "" {
		items = append(items, ConfirmationItem{
			Label: "説明",
			Value: config.Description,
		})
	} else {
		items = append(items, ConfirmationItem{
			Label: "説明",
			Value: "（なし）",
		})
	}

	items = append(items, ConfirmationItem{
		Label: "README追加",
		Value: func() string {
			if config.AddReadme {
				return "✅ はい"
			}
			return "❌ いいえ"
		}(),
	})

	items = append(items, ConfirmationItem{
		Label: "作成後にクローン",
		Value: func() string {
			if config.SholdClone {
				return "✅ はい"
			}
			return "❌ いいえ"
		}(),
	})

	return ConfirmationSection{
		Title: "リポジトリ設定",
		Icon:  "⚙️",
		Items: items,
	}
}

// buildDestinationSection は作成先情報セクションを構築する
func buildDestinationSection(state *WizardState, githubClient interface{}) ConfirmationSection {
	user := getCurrentUserWithClient(githubClient)
	url := "（未設定）"

	if state.RepoConfig != nil && state.RepoConfig.Name != "" {
		url = fmt.Sprintf("https://github.com/%s/%s", user, state.RepoConfig.Name)
	}

	items := []ConfirmationItem{
		{
			Label: "GitHubユーザー",
			Value: user,
		},
		{
			Label:     "作成先URL",
			Value:     url,
			Important: true,
		},
	}

	return ConfirmationSection{
		Title: "作成先",
		Icon:  "📍",
		Items: items,
	}
}

// buildWarnings は警告メッセージを構築する
func buildWarnings(state *WizardState) []string {
	var warnings []string

	// リポジトリ名の重複チェック（簡易版）
	if state.RepoConfig != nil {
		if strings.Contains(strings.ToLower(state.RepoConfig.Name), "test") {
			warnings = append(warnings, "リポジトリ名に「test」が含まれています。本番用の場合は変更を検討してください。")
		}
	}

	// パブリックリポジトリの警告
	if state.RepoConfig != nil && !state.RepoConfig.IsPrivate {
		warnings = append(warnings, "パブリックリポジトリは全世界に公開されます。機密情報が含まれていないか確認してください。")
	}

	// テンプレート使用時の警告
	if state.UseTemplate && state.SelectedTemplate != nil {
		if time.Since(state.SelectedTemplate.UpdatedAt) > 365*24*time.Hour {
			warnings = append(warnings, "選択されたテンプレートは1年以上更新されていません。最新の状況を確認することをお勧めします。")
		}
	}

	return warnings
}

// getCurrentUser は現在のGitHubユーザー名を取得する（後方互換用）
func getCurrentUser() string {
	return getCurrentUserWithClient(nil)
}

// getCurrentUserWithClient は現在のGitHubユーザー名を取得する
func getCurrentUserWithClient(githubClient interface{}) string {
	// デバッグ出力を追加
	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: getCurrentUserWithClient called, client type: %T\n", githubClient)
	}
	
	// もしクライアントがnilの場合は、直接GitHubCLIから取得を試行
	if githubClient == nil {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Client is nil, trying direct approach\n")
		}
		return "your-username"
	}
	
	// より汎用的な型アサーション - どんな構造体でも試行
	if userInterface := tryGetCurrentUserInterface(githubClient); userInterface != nil {
		if login := extractLoginFromUser(userInterface); login != "" {
			if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "DEBUG: Extracted login: %s\n", login)
			}
			return login
		}
	}

	// フォールバック: 簡易実装
	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: Fallback to your-username\n")
	}
	return "your-username"
}

// githubUser はGitHubユーザーを表すローカル型
type githubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// tryGetCurrentUserInterface は任意の型のGetCurrentUserメソッドを呼び出す  
func tryGetCurrentUserInterface(client interface{}) interface{} {
	// デバッグ出力
	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: tryGetCurrentUserInterface called\n")
	}
	
	// github.User型に対応した型アサーション
	type UserWithLogin struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	
	type getCurrentUserMethod interface {
		GetCurrentUser() (*UserWithLogin, error)
	}

	if c, ok := client.(getCurrentUserMethod); ok {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Type assertion for GetCurrentUser with UserWithLogin succeeded\n")
		}
		if user, err := c.GetCurrentUser(); err == nil {
			if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "DEBUG: GetCurrentUser() returned: %T = %+v\n", user, user)
			}
			return user
		} else {
			if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "DEBUG: GetCurrentUser() error: %v\n", err)
			}
		}
	} else {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Type assertion for GetCurrentUser with UserWithLogin failed\n")
		}
	}
	
	// フォールバック：より汎用的な型
	type genericGetCurrentUserMethod interface {
		GetCurrentUser() (interface{}, error)
	}

	if c, ok := client.(genericGetCurrentUserMethod); ok {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Type assertion for generic GetCurrentUser succeeded\n")
		}
		if user, err := c.GetCurrentUser(); err == nil {
			if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "DEBUG: Generic GetCurrentUser() returned: %T = %+v\n", user, user)
			}
			return user
		} else {
			if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer debugFile.Close()
				fmt.Fprintf(debugFile, "DEBUG: Generic GetCurrentUser() error: %v\n", err)
			}
		}
	} else {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Type assertion for generic GetCurrentUser failed\n")
		}
	}

	return nil
}

// extractLoginFromUser はユーザーオブジェクトからログイン名を抽出
func extractLoginFromUser(user interface{}) string {
	// デバッグ出力
	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: extractLoginFromUser called, user type: %T, value: %+v\n", user, user)
	}
	
	// Loginフィールドを持つ任意の型に対応（リフレクション代わり）
	type LoginProvider interface {
		GetLogin() string
	}
	
	if loginProvider, ok := user.(LoginProvider); ok {
		login := loginProvider.GetLogin()
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Found LoginProvider with Login: %s\n", login)
		}
		return login
	}
	
	// 実際のgithub.User型のパターン（ポインタ型） - より多くのバリエーションを試行
	type UserType1 struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	
	if u, ok := user.(*UserType1); ok && u != nil {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Found UserType1 pointer with Login: %s\n", u.Login)
		}
		return u.Login
	}
	
	// 直接的な型アサーション - runtime type information の確認
	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: Trying to match actual runtime type: %T\n", user)
	}
	
	// github.User構造体のパターン（値型）
	if u, ok := user.(struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}); ok {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Found value struct with Login: %s\n", u.Login)
		}
		return u.Login
	}

	// 一般的な構造体パターンを試行
	if u, ok := user.(struct {
		Login     string
		Name      string
		Email     string
		AvatarURL string
	}); ok {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Found general struct with Login: %s\n", u.Login)
		}
		return u.Login
	}

	// シンプルなLoginフィールドのみ
	if u, ok := user.(struct {
		Login string
	}); ok {
		if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			defer debugFile.Close()
			fmt.Fprintf(debugFile, "DEBUG: Found simple Login struct: %s\n", u.Login)
		}
		return u.Login
	}

	if debugFile, err := os.OpenFile("/tmp/gh-wizard-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "DEBUG: No matching type found for extracting Login\n")
	}
	return ""
}

// GetActionByKey はキー入力からアクションを取得する
func GetActionByKey(key string) (ConfirmationAction, bool) {
	actions := []ConfirmationAction{
		ActionModifySettings,
		ActionCancel,
		ActionCreateRepository,
	}

	for _, action := range actions {
		if action.GetKey() == key {
			return action, true
		}
	}

	return ActionModifySettings, false
}

// FormatRepositoryCommand はリポジトリ作成コマンドを整形して返す
func (cd *ConfirmationData) FormatRepositoryCommand(state *WizardState) []string {
	if state.RepoConfig == nil {
		return []string{}
	}

	return state.RepoConfig.GetGHCommand(state.SelectedTemplate)
}
