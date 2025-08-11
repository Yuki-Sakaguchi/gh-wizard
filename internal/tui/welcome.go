package tui

import (
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WelcomeView はウェルカム画面のモデル
type WelcomeView struct {
	state    *models.WizardState
	styles   *Styles
	width    int
	height   int
	selected int // 選択中のオプション (0: テンプレート使用, 1: 空のリポジトリ)
}

// NewWelcomeView は新しいウェルカム画面を作成する
func NewWelcomeView(state *models.WizardState, styles *Styles) *WelcomeView {
	return &WelcomeView{
		state:    state,
		styles:   styles,
		selected: 0,
	}
}

// Init は Bubble Tea の初期化
func (v *WelcomeView) Init() tea.Cmd {
	return nil
}

// Update は Bubble Tea のアップデート処理
func (v *WelcomeView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.selected > 0 {
				v.selected--
			}
		case "down", "j":
			if v.selected < 1 {
				v.selected++
			}
		case "enter":
			// 選択を確定
			v.state.UseTemplate = (v.selected == 0)
			return v, func() tea.Msg {
				if v.state.UseTemplate {
					return StepChangeMsg{Step: models.StepTemplateSelection}
				} else {
					return StepChangeMsg{Step: models.StepRepositorySettings}
				}
			}
		}
	}

	return v, nil
}

// View は Bubble Tea のビュー描画
func (v *WelcomeView) View() string {
	if v.width == 0 {
		return "読み込み中..."
	}

	// タイトル
	title := v.styles.Title.Render("🔮 GitHub Repository Wizard")

	// サブタイトル
	subtitle := v.styles.Subtitle.Render("GitHubリポジトリを魔法のように簡単に作成します")

	// 質問
	question := v.styles.Text.Render("テンプレートリポジトリを使用しますか？")

	// オプション
	options := []string{
		"はい - 既存のテンプレートから作成",
		"いいえ - 空のリポジトリを作成",
	}

	var optionViews []string
	for i, option := range options {
		style := v.styles.Unselected
		prefix := "  "

		if i == v.selected {
			style = v.styles.Selected
			prefix = "▸ "
		}

		optionViews = append(optionViews, prefix+style.Render(option))
	}

	// キーバインドヘルプ
	help := v.styles.Info.Render("⌨️  ↑↓: 選択  Enter: 決定  q: 終了")

	// レイアウト
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		subtitle,
		"",
		"",
		question,
		"",
		strings.Join(optionViews, "\n"),
		"",
		"",
		help,
	)

	// 中央寄せ
	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		v.styles.Border.Render(content),
	)
}

// SetSize はビューのサイズを設定する
func (v *WelcomeView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// GetTitle はビューのタイトルを返す
func (v *WelcomeView) GetTitle() string {
	return "ウェルカム"
}

// CanGoBack は前に戻れるかを返す
func (v *WelcomeView) CanGoBack() bool {
	return false // ウェルカム画面は最初の画面
}

// CanGoNext は次に進めるかを返す
func (v *WelcomeView) CanGoNext() bool {
	return true
}
