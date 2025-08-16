package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExecutionView は実行画面のビューコントローラー
type ExecutionView struct {
	state        *models.WizardState
	styles       *Styles
	githubClient github.Client

	// UI コンポーネント
	spinner      spinner.Model
	progress     progress.Model
	taskProgress map[string]progress.Model

	// 実行状態
	plan         *models.ExecutionPlan
	progressChan chan models.ExecutionMessage
	isExecuting  bool
	isCompleted  bool
	executionErr error
	result       *models.RepositoryCreationResult

	// レイアウト情報
	width  int
	height int
}

// NewExecutionView は新しい実行画面を作成する
func NewExecutionView(state *models.WizardState, styles *Styles, githubClient github.Client) *ExecutionView {
	// スピナーの設定
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Colors.Accent))

	// プログレスバーの設定
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 50

	return &ExecutionView{
		state:        state,
		styles:       styles,
		githubClient: githubClient,
		spinner:      s,
		progress:     p,
		taskProgress: make(map[string]progress.Model),
		progressChan: make(chan models.ExecutionMessage, 100),
	}
}

// Init は初期化コマンドを返す
func (v *ExecutionView) Init() tea.Cmd {
	v.plan = models.NewExecutionPlan(v.state)

	// 各タスク用のプログレスバーを作成
	for _, task := range v.plan.Tasks {
		taskProgress := progress.New(progress.WithDefaultGradient())
		taskProgress.Width = 30
		v.taskProgress[task.ID] = taskProgress
	}

	return tea.Batch(
		v.spinner.Tick,
		v.startExecution(),
		v.listenForProgress(),
	)
}

// Update は Bubble Tea のアップデート処理
func (v *ExecutionView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if v.isCompleted {
				// 完了画面に進む（結果を含む）
				return v, StepChangeWithResultCmd(models.StepCompleted, v.result)
			}
		}

	case ExecutionCompleteMsg:
		v.isCompleted = true
		v.isExecuting = false
		return v, nil

	case ExecutionErrorMsg:
		v.executionErr = msg.Error
		v.isCompleted = true
		v.isExecuting = false
		return v, nil

	case ExecutionProgressMsg:
		// プログレス更新
		v.updateProgress(msg.Message)
		// 継続的にプログレスをリッスン
		return v, v.listenForProgress()

	case ExecutionTickMsg:
		// プログレス継続監視
		return v, v.listenForProgress()

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progModel, cmd := v.progress.Update(msg)
		if p, ok := progModel.(progress.Model); ok {
			v.progress = p
		}
		cmds = append(cmds, cmd)

		// 各タスクのプログレスも更新
		for id, taskProg := range v.taskProgress {
			updatedModel, taskCmd := taskProg.Update(msg)
			if tp, ok := updatedModel.(progress.Model); ok {
				v.taskProgress[id] = tp
			}
			cmds = append(cmds, taskCmd)
		}
	}

	return v, tea.Batch(cmds...)
}

// View はビューを描画する
func (v *ExecutionView) View() string {
	var b strings.Builder

	// タイトル
	title := v.styles.Title.Render("リポジトリを作成中...")
	b.WriteString(title + "\n\n")

	// 全体の進捗バー
	overallProgress := "全体進捗: "
	if v.plan != nil {
		percentage := int(v.plan.OverallProgress * 100)
		progressBar := v.progress.ViewAs(v.plan.OverallProgress)
		overallProgress += fmt.Sprintf("%s %d%%", progressBar, percentage)

		// 推定残り時間
		if v.isExecuting {
			remaining := v.plan.GetEstimatedRemainingTime()
			if remaining > 0 {
				overallProgress += fmt.Sprintf(" (残り約 %s)", formatDuration(remaining))
			}
		}
	}
	b.WriteString(overallProgress + "\n\n")

	// タスク詳細
	if v.plan != nil {
		b.WriteString(v.renderTaskList())
	}

	// ステータスメッセージ
	b.WriteString("\n\n")
	if v.executionErr != nil {
		errorMsg := v.styles.Error.Render(fmt.Sprintf("エラーが発生しました: %v", v.executionErr))
		b.WriteString(errorMsg + "\n")
		b.WriteString(v.styles.Info.Render("Enterキーで完了画面に進みます"))
	} else if v.isCompleted {
		successMsg := v.styles.Success.Render("リポジトリの作成が完了しました！")
		b.WriteString(successMsg + "\n")
		b.WriteString(v.styles.Info.Render("Enterキーで完了画面に進みます"))
	} else if v.isExecuting {
		spinner := v.spinner.View()
		b.WriteString(v.styles.Info.Render(fmt.Sprintf("%s 処理中...", spinner)))
	}

	// 全体をボーダーで囲む
	content := b.String()
	return v.styles.Border.
		Width(v.width - 4).
		Height(v.height - 4).
		Render(content)
}

// renderTaskList はタスクリストを描画する
func (v *ExecutionView) renderTaskList() string {
	var b strings.Builder

	b.WriteString(v.styles.Subtitle.Render("実行タスク:") + "\n")

	for _, task := range v.plan.Tasks {
		// タスク状態アイコン
		icon := task.Status.GetIcon()

		// タスク名とステータス
		taskLine := fmt.Sprintf("%s %s", icon, task.Name)

		// プログレスバー（実行中の場合）
		if task.Status == models.TaskStatusInProgress && task.Progress > 0 {
			progressBar := v.taskProgress[task.ID].ViewAs(task.Progress)
			percentage := int(task.Progress * 100)
			taskLine += fmt.Sprintf(" %s %d%%", progressBar, percentage)
		}

		// 経過時間（実行中または完了の場合）
		if !task.StartTime.IsZero() {
			elapsed := task.GetElapsedTime()
			if elapsed > 0 {
				taskLine += fmt.Sprintf(" (%s)", formatDuration(elapsed))
			}
		}

		// スタイル適用
		switch task.Status {
		case models.TaskStatusCompleted:
			taskLine = v.styles.Success.Render(taskLine)
		case models.TaskStatusFailed:
			taskLine = v.styles.Error.Render(taskLine)
		case models.TaskStatusInProgress:
			accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(v.styles.Colors.Accent)).Bold(true)
			taskLine = accentStyle.Render(taskLine)
		case models.TaskStatusSkipped:
			taskLine = v.styles.Debug.Render(taskLine)
		default:
			taskLine = v.styles.Text.Render(taskLine)
		}

		b.WriteString("  " + taskLine + "\n")

		// エラーメッセージを表示
		if task.Status == models.TaskStatusFailed && task.Error != nil {
			errorLine := v.styles.Error.Render(fmt.Sprintf("    エラー: %v", task.Error))
			b.WriteString(errorLine + "\n")
		}
	}

	return b.String()
}

// updateProgress はプログレス情報を更新する
func (v *ExecutionView) updateProgress(message models.ExecutionMessage) {
	if v.plan == nil {
		return
	}

	v.plan.UpdateTaskStatus(message.TaskID, message.Status, message.Progress, message.Error)

	// 現在のタスクIDを更新
	if message.Status == models.TaskStatusInProgress {
		v.plan.CurrentTaskID = message.TaskID
	}
}

// startExecution は実行を開始する
func (v *ExecutionView) startExecution() tea.Cmd {
	return func() tea.Msg {
		v.isExecuting = true

		go func() {
			ctx := context.Background()
			result, err := v.githubClient.CreateRepositoryWithProgress(ctx, v.state, v.progressChan)

			// 結果を保存
			v.result = result

			if err != nil {
				v.progressChan <- models.ExecutionMessage{
					TaskID:  "execution",
					Status:  models.TaskStatusFailed,
					Error:   err,
					Message: "実行が失敗しました",
				}
			} else {
				v.progressChan <- models.ExecutionMessage{
					TaskID:  "execution",
					Status:  models.TaskStatusCompleted,
					Message: "すべての処理が完了しました",
				}
			}
			close(v.progressChan)
		}()

		return nil
	}
}

// listenForProgress はプログレス更新をリッスンする
func (v *ExecutionView) listenForProgress() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg, ok := <-v.progressChan:
			if !ok {
				// チャネルがクローズされた = 実行完了
				if v.executionErr == nil {
					return ExecutionCompleteMsg{}
				}
				return ExecutionErrorMsg{Error: v.executionErr}
			}

			// エラーメッセージの場合
			if msg.Error != nil {
				v.executionErr = msg.Error
				return ExecutionErrorMsg{Error: msg.Error}
			}

			return ExecutionProgressMsg{Message: msg}

		default:
			// ノンブロッキングでチェックして、メッセージがなければ継続
			return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
				return ExecutionTickMsg{}
			})()
		}
	}
}

// SetSize はビューのサイズを設定する
func (v *ExecutionView) SetSize(width, height int) {
	v.width = width
	v.height = height

	// プログレスバーの幅を調整
	maxProgressWidth := width - 20
	if maxProgressWidth > 0 {
		v.progress.Width = min(maxProgressWidth, 60)

		taskProgressWidth := min(maxProgressWidth-20, 40)
		for id, prog := range v.taskProgress {
			prog.Width = taskProgressWidth
			v.taskProgress[id] = prog
		}
	}
}

// GetTitle はタイトルを返す
func (v *ExecutionView) GetTitle() string {
	return "実行中"
}

// CanGoBack は前に戻れるかを返す
func (v *ExecutionView) CanGoBack() bool {
	return false // 実行中は戻れない
}

// CanGoNext は次に進めるかを返す
func (v *ExecutionView) CanGoNext() bool {
	return v.isCompleted
}

// カスタムメッセージ型
type ExecutionCompleteMsg struct{}

type ExecutionErrorMsg struct {
	Error error
}

type ExecutionProgressMsg struct {
	Message models.ExecutionMessage
}

type ExecutionTickMsg struct{}

// StepChangeCmd はステップ変更コマンドを作成する
func StepChangeCmd(step models.Step) tea.Cmd {
	return func() tea.Msg {
		return StepChangeMsg{Step: step}
	}
}

// StepChangeWithResultCmd は結果付きステップ変更コマンドを作成する
func StepChangeWithResultCmd(step models.Step, result *models.RepositoryCreationResult) tea.Cmd {
	return func() tea.Msg {
		return models.StepChangeWithResultMsg{Step: step, Result: result}
	}
}

// ヘルパー関数
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
