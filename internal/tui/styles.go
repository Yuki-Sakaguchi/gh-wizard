package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles は TUI のスタイル定義
type Styles struct {
	// 基本スタイル
	Base     lipgloss.Style
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Text     lipgloss.Style

	// UI要素
	Border     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Focused    lipgloss.Style
	Blurred    lipgloss.Style

	// ステータス
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style

	// デバッグ
	Debug lipgloss.Style

	// カラーパレット
	Colors ColorPalette
}

// ColorPalette は色の定義
type ColorPalette struct {
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Foreground string
	Success    string
	Warning    string
	Error      string
	Info       string
	Debug      string
}

// NewStyles は新しいスタイルセットを作成する
func NewStyles() *Styles {
	colors := ColorPalette{
		Primary:    "#6366f1", // インディゴ
		Secondary:  "#8b5cf6", // バイオレット
		Accent:     "#06b6d4", // シアン
		Background: "#1f2937", // グレー800
		Foreground: "#f9fafb", // グレー50
		Success:    "#10b981", // エメラルド
		Warning:    "#f59e0b", // アンバー
		Error:      "#ef4444", // レッド
		Info:       "#3b82f6", // ブルー
		Debug:      "#6b7280", // グレー500
	}

	return &Styles{
		Colors: colors,

		// 基本スタイル
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Foreground)),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Primary)).
			Bold(true).
			Padding(0, 1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Secondary)).
			Italic(true),

		Text: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Foreground)),

		// UI要素
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Primary)).
			Padding(1, 2),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Background)).
			Background(lipgloss.Color(colors.Primary)).
			Bold(true).
			Padding(0, 1),

		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Foreground)).
			Padding(0, 1),

		Focused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Primary)),

		Blurred: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colors.Debug)),

		// ステータス
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Success)).
			Bold(true),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Warning)).
			Bold(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Error)).
			Bold(true),

		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Info)),

		Debug: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Debug)).
			Italic(true),
	}
}

// Width はスタイルに幅を設定する
func (s *Styles) Width(width int) *Styles {
	newStyles := *s
	newStyles.Border = s.Border.Width(width - 4)
	return &newStyles
}

// Height はスタイルに高さを設定する
func (s *Styles) Height(height int) *Styles {
	newStyles := *s
	newStyles.Border = s.Border.Height(height - 4)
	return &newStyles
}
