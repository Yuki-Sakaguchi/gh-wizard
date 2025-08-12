package tui

import (
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/config"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// App は TUI アプリケーションのメインモデル
type App struct {
	width  int
	height int

	// 依存関係
	config       *config.Config
	githubClient github.Client

	// 状態管理
	state       *models.WizardState
	currentView ViewController

	// UI コンポーネント
	styles *Styles

	// デバッグ用
	debug        bool
	debugMessage string
}

// ViewController は各画面のインターフェース
type ViewController interface {
	Init() tea.Cmd
	Update(tea.Msg) (ViewController, tea.Cmd)
	View() string

	SetSize(width, height int)
	GetTitle() string
	CanGoBack() bool
	CanGoNext() bool
}

// NewApp は新しい TUI アプリケーションを作成する
func NewApp(cfg *config.Config, githubClient github.Client, debug bool) *App {
	app := &App{
		config:       cfg,
		githubClient: githubClient,
		state:        models.NewWizardState(),
		styles:       NewStyles(),
		debug:        debug,
	}

	// 初期画面を設定
	app.currentView = NewWelcomeView(app.state, app.styles)

	return app
}

// Init は Bubble Tea の初期化
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.currentView.Init(),
		tea.EnterAltScreen, // フルスクリーンモード
	)
}

// Update は Bubble Tea のアップデート処理
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.currentView.SetSize(a.width, a.height)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// アプリケーション終了
			return a, tea.Quit

		case "ctrl+d":
			a.debug = !a.debug
			a.debugMessage = fmt.Sprintf("Debug mode: %v", a.debug)

		case "esc":
			// 前の画面に戻る
			if a.currentView.CanGoBack() && a.state.CurrentStep > models.StepWelcome {
				return a.navigateBack()
			}

		case "enter":
			// 次の画面に進む
			if a.currentView.CanGoNext() && a.state.CanProceedToNext() {
				return a.navigateNext()
			}
		}

	case StepChangeMsg:
		// ステップ変更メッセージ
		return a.changeStep(msg.Step)
		
	case models.StepChangeWithResultMsg:
		// 結果付きステップ変更メッセージ
		return a.changeStepWithResult(msg.Step, msg.Result)

	case ErrorMsg:
		a.debugMessage = fmt.Sprintf("Error: %v", msg.Error)
	}

	// 現在の画面にメッセージを転送
	a.currentView, cmd = a.currentView.Update(msg)

	return a, cmd
}

// View は Bubble Tea のビューの描画
func (a *App) View() string {
	if a.width == 0 {
		return "初期化中..."
	}

	// メインコンテンツ
	content := a.currentView.View()

	// デバッグ情報を追加
	if a.debug && a.debugMessage != "" {
		content += "\n\n" + a.styles.Debug.Render("DEBUG:"+a.debugMessage)
	}

	return content
}

// navigateNext は次のステップに進む
func (a *App) navigateNext() (tea.Model, tea.Cmd) {
	nextStep := a.state.CurrentStep + 1
	return a.changeStep(nextStep)
}

// navigateBack は前のステップに戻る
func (a *App) navigateBack() (tea.Model, tea.Cmd) {
	prevStep := a.state.CurrentStep - 1
	return a.changeStep(prevStep)
}

// changeStep はステップを変更し、対応する画面を表示する
func (a *App) changeStep(step models.Step) (tea.Model, tea.Cmd) {
	a.state.CurrentStep = step

	switch step {
	case models.StepWelcome:
		a.currentView = NewWelcomeView(a.state, a.styles)

	case models.StepTemplateSelection:
		a.currentView = NewTemplateView(a.state, a.styles, a.githubClient)

	case models.StepRepositorySettings:
		a.currentView = NewSettingsView(a.state, a.styles)

	case models.StepConfirmation:
		a.currentView = NewConfirmationView(a.state, a.styles, a.githubClient)

	case models.StepExecution:
		a.currentView = NewExecutionView(a.state, a.styles, a.githubClient)

	case models.StepCompleted:
		// 完了画面用の結果データを作成
		result := &models.RepositoryCreationResult{
			Success:       true,
			RepositoryURL: fmt.Sprintf("https://github.com/your-username/%s", a.state.RepoConfig.Name),
			Message:       "リポジトリが正常に作成されました",
		}
		if a.state.RepoConfig.SholdClone {
			result.ClonePath = fmt.Sprintf("./%s", a.state.RepoConfig.Name)
		}
		a.currentView = NewCompletedView(a.state, a.styles, result)

	default:
		a.debugMessage = fmt.Sprintf("Unknown step: %v", step)
		return a, nil
	}

	a.currentView.SetSize(a.width, a.height)

	return a, a.currentView.Init()
}

// changeStepWithResult は結果付きでステップを変更し、対応する画面を表示する
func (a *App) changeStepWithResult(step models.Step, result *models.RepositoryCreationResult) (tea.Model, tea.Cmd) {
	a.state.CurrentStep = step

	switch step {
	case models.StepCompleted:
		a.currentView = NewCompletedView(a.state, a.styles, result)
	default:
		// 他のステップは通常のchangeStepで処理
		return a.changeStep(step)
	}

	a.currentView.SetSize(a.width, a.height)
	return a, a.currentView.Init()
}

// カスタムメッセージ型
type StepChangeMsg struct {
	Step models.Step
}

type ErrorMsg struct {
	Error error
}

// Run は TUI アプリケーションを実行する
func Run(cfg *config.Config, githubClient github.Client, debug bool) error {
	app := NewApp(cfg, githubClient, debug)

	// Bubble Tea プログラムを作成・実装
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// プログラムを実行
	_, err := p.Run()
	return err
}
