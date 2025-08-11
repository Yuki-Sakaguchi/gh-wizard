package tui

import (
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmationView は確認画面
type ConfirmationView struct {
	state  *models.WizardState
	styles *Styles
	width  int
	height int
}

func NewConfirmationView(state *models.WizardState, styles *Styles) *ConfirmationView {
	return &ConfirmationView{
		state:  state,
		styles: styles,
	}
}

func (v *ConfirmationView) Init() tea.Cmd                                { return nil }
func (v *ConfirmationView) Update(msg tea.Msg) (ViewController, tea.Cmd) { return v, nil }
func (v *ConfirmationView) View() string                                 { return "リポジトリ設定画面（未実装）" }
func (v *ConfirmationView) SetSize(width, height int)                    { v.width, v.height = width, height }
func (v *ConfirmationView) GetTitle() string                             { return "リポジトリ設定" }
func (v *ConfirmationView) CanGoBack() bool                              { return true }
func (v *ConfirmationView) CanGoNext() bool                              { return true }
