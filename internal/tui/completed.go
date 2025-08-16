package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// CompletedView は完了画面のビューコントローラー
type CompletedView struct {
	state  *models.WizardState
	styles *Styles

	// 完了情報
	result         *models.RepositoryCreationResult
	completionTime time.Time

	// レイアウト情報
	width  int
	height int
}

// NewCompletedView は新しい完了画面を作成する
func NewCompletedView(state *models.WizardState, styles *Styles, result *models.RepositoryCreationResult) *CompletedView {
	return &CompletedView{
		state:          state,
		styles:         styles,
		result:         result,
		completionTime: time.Now(),
	}
}

// Init は初期化コマンドを返す
func (v *CompletedView) Init() tea.Cmd {
	return nil
}

// Update は Bubble Tea のアップデート処理
func (v *CompletedView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return v, tea.Quit
		case "enter", "esc":
			// アプリケーションを終了するかウェルカム画面に戻る
			return v, tea.Quit
		case "r", "R":
			// 新しいリポジトリ作成を開始
			return v, StepChangeCmd(models.StepWelcome)
		}
	}

	return v, nil
}

// View はビューを描画する
func (v *CompletedView) View() string {
	var b strings.Builder

	// タイトル
	var title string
	if v.result != nil && v.result.Success {
		title = "🎉 リポジトリの作成が完了しました！"
	} else {
		title = "❌ リポジトリの作成に失敗しました"
	}

	styledTitle := v.styles.Title.Render(title)
	b.WriteString(styledTitle + "\n\n")

	// 結果詳細
	if v.result != nil {
		resultSection := v.renderResultSection()
		b.WriteString(resultSection + "\n")
	}

	// リポジトリ情報
	if v.result != nil && v.result.Success {
		repoSection := v.renderRepositoryInfo()
		b.WriteString(repoSection + "\n")
	}

	// 次のステップ
	nextStepsSection := v.renderNextSteps()
	b.WriteString(nextStepsSection + "\n")

	// アクション
	actionsSection := v.renderActions()
	b.WriteString(actionsSection + "\n")

	// 全体をボーダーで囲む
	content := b.String()
	return v.styles.Border.
		Width(v.width - 4).
		Height(v.height - 4).
		Render(content)
}

// renderResultSection は結果セクションを描画する
func (v *CompletedView) renderResultSection() string {
	var lines []string

	// セクションタイトル
	sectionTitle := v.styles.Subtitle.Render("📊 実行結果")
	lines = append(lines, sectionTitle)

	// ステータス
	var statusLine string
	if v.result.Success {
		statusLine = v.styles.Success.Render("✅ 成功: " + v.result.Message)
	} else {
		statusLine = v.styles.Error.Render("❌ 失敗: " + v.result.Message)
		if v.result.Error != nil {
			errorDetail := v.styles.Error.Render(fmt.Sprintf("   エラー詳細: %v", v.result.Error))
			lines = append(lines, errorDetail)
		}
	}
	lines = append(lines, statusLine)

	// 完了時刻
	completionTime := v.styles.Info.Render(fmt.Sprintf("⏰ 完了時刻: %s", v.completionTime.Format("2006-01-02 15:04:05")))
	lines = append(lines, completionTime)

	return strings.Join(lines, "\n") + "\n"
}

// renderRepositoryInfo はリポジトリ情報を描画する
func (v *CompletedView) renderRepositoryInfo() string {
	if !v.result.Success {
		return ""
	}

	var lines []string

	// セクションタイトル
	sectionTitle := v.styles.Subtitle.Render("📁 リポジトリ情報")
	lines = append(lines, sectionTitle)

	// リポジトリURL
	if v.result.RepositoryURL != "" {
		repoURL := v.styles.Info.Render(fmt.Sprintf("🔗 URL: %s", v.result.RepositoryURL))
		lines = append(lines, repoURL)
	}

	// クローンパス
	if v.result.ClonePath != "" {
		clonePath := v.styles.Info.Render(fmt.Sprintf("📂 ローカルパス: %s", v.result.ClonePath))
		lines = append(lines, clonePath)
	}

	// リポジトリ設定の詳細
	if v.state.RepoConfig != nil {
		config := v.state.RepoConfig

		repoName := v.styles.Text.Render(fmt.Sprintf("   名前: %s", config.Name))
		lines = append(lines, repoName)

		if config.Description != "" {
			repoDesc := v.styles.Text.Render(fmt.Sprintf("   説明: %s", config.Description))
			lines = append(lines, repoDesc)
		}

		var features []string
		if config.IsPrivate {
			features = append(features, "プライベート")
		} else {
			features = append(features, "パブリック")
		}
		if config.AddReadme {
			features = append(features, "README.md")
		}
		if config.SholdClone {
			features = append(features, "ローカルクローン")
		}
		if len(features) > 0 {
			featuresLine := v.styles.Text.Render(fmt.Sprintf("   機能: %s", strings.Join(features, ", ")))
			lines = append(lines, featuresLine)
		}
	}

	// テンプレート情報
	if v.state.UseTemplate && v.state.SelectedTemplate != nil {
		template := v.state.SelectedTemplate
		templateLine := v.styles.Text.Render(fmt.Sprintf("   テンプレート: %s", template.FullName))
		lines = append(lines, templateLine)
	}

	return strings.Join(lines, "\n") + "\n"
}

// renderNextSteps は次のステップを描画する
func (v *CompletedView) renderNextSteps() string {
	var lines []string

	// セクションタイトル
	sectionTitle := v.styles.Subtitle.Render("🚀 次のステップ")
	lines = append(lines, sectionTitle)

	if v.result != nil && v.result.Success {
		// 成功時の次のステップ
		steps := []string{
			"リポジトリを開いて開発を始める",
			"チームメンバーをコラボレーターとして追加",
			"プロジェクトの設定とCI/CDを構成",
		}

		if v.result.ClonePath != "" {
			steps[0] = fmt.Sprintf("cd %s でローカルディレクトリに移動", v.result.ClonePath)
		}

		for i, step := range steps {
			stepLine := v.styles.Text.Render(fmt.Sprintf("%d. %s", i+1, step))
			lines = append(lines, stepLine)
		}

		// 便利なコマンド
		lines = append(lines, "")
		commandsTitle := v.styles.Debug.Render("💡 便利なコマンド:")
		lines = append(lines, commandsTitle)

		commands := []string{
			fmt.Sprintf("gh repo view %s/%s", "your-username", v.state.RepoConfig.Name),
		}
		if v.result.ClonePath != "" {
			commands = append(commands, fmt.Sprintf("cd %s", v.result.ClonePath))
		}
		commands = append(commands, "git status")

		for _, cmd := range commands {
			cmdLine := v.styles.Debug.Render(fmt.Sprintf("  $ %s", cmd))
			lines = append(lines, cmdLine)
		}
	} else {
		// 失敗時の次のステップ
		steps := []string{
			"エラー内容を確認して設定を見直す",
			"GitHub CLI の認証状態を確認 (gh auth status)",
			"必要に応じてマニュアルでリポジトリを作成",
		}

		for i, step := range steps {
			stepLine := v.styles.Warning.Render(fmt.Sprintf("%d. %s", i+1, step))
			lines = append(lines, stepLine)
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

// renderActions はアクションボタンを描画する
func (v *CompletedView) renderActions() string {
	var lines []string

	// アクション説明
	actionTitle := v.styles.Subtitle.Render("⌨️  操作")
	lines = append(lines, actionTitle)

	// アクションリスト
	actions := []string{
		"Enter / Esc: アプリケーションを終了",
		"R: 新しいリポジトリの作成を開始",
		"Q / Ctrl+C: 強制終了",
	}

	for _, action := range actions {
		actionLine := v.styles.Info.Render(action)
		lines = append(lines, actionLine)
	}

	return strings.Join(lines, "\n")
}

// SetSize はビューのサイズを設定する
func (v *CompletedView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// GetTitle はタイトルを返す
func (v *CompletedView) GetTitle() string {
	if v.result != nil && v.result.Success {
		return "完了"
	}
	return "失敗"
}

// CanGoBack は前に戻れるかを返す
func (v *CompletedView) CanGoBack() bool {
	return false // 完了画面からは戻れない
}

// CanGoNext は次に進めるかを返す
func (v *CompletedView) CanGoNext() bool {
	return false // 完了画面が最終
}
