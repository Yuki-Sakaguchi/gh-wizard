package tui

import (
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsView はリポジトリ設定画面（プレースホルダー）
type SettingsView struct {
	state  *models.WizardState
	styles *Styles
	width  int
	height int
}

func NewSettingsView(state *models.WizardState, styles *Styles) *SettingsView {
	return &SettingsView{
		state:  state,
		styles: styles,
	}
}

func (v *SettingsView) Init() tea.Cmd                                { return nil }
func (v *SettingsView) Update(msg tea.Msg) (ViewController, tea.Cmd) { return v, nil }
func (v *SettingsView) View() string                                 { return "リポジトリ設定画面（未実装）" }
func (v *SettingsView) SetSize(width, height int)                    { v.width, v.height = width, height }
func (v *SettingsView) GetTitle() string                             { return "リポジトリ設定" }
func (v *SettingsView) CanGoBack() bool                              { return true }
func (v *SettingsView) CanGoNext() bool                              { return true }
