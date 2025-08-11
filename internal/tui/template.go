package tui

import (
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// TemplateView はテンプレート選択画面（プレースホルダー）
type TemplateView struct {
	state        *models.WizardState
	styles       *Styles
	githubClient github.Client
	width        int
	height       int
}

func NewTemplateView(state *models.WizardState, styles *Styles, githubClient github.Client) *TemplateView {
	return &TemplateView{
		state:        state,
		styles:       styles,
		githubClient: githubClient,
	}
}

func (v *TemplateView) Init() tea.Cmd                                { return nil }
func (v *TemplateView) Update(msg tea.Msg) (ViewController, tea.Cmd) { return v, nil }
func (v *TemplateView) View() string                                 { return "テンプレート選択画面（未実装）" }
func (v *TemplateView) SetSize(width, height int)                    { v.width, v.height = width, height }
func (v *TemplateView) GetTitle() string                             { return "テンプレート選択" }
func (v *TemplateView) CanGoBack() bool                              { return true }
func (v *TemplateView) CanGoNext() bool                              { return true }
