package tui

import (
	"fmt"
	"strings"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsView はリポジトリ設定画面（連続入力形式）
type SettingsView struct {
	state  *models.WizardState
	styles *Styles
	width  int
	height int

	// 質問フロー管理
	questionFlow *models.QuestionFlow

	// UI コンポーネント
	textInput   textinput.Model
	progress    progress.Model
	selectIndex int // Select型質問用のインデックス

	// 状態管理
	inputValue   string
	errorMessage string
	showHelp     bool
}

func NewSettingsView(state *models.WizardState, styles *Styles) *SettingsView {
	// TextInputの初期化
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	// Progressの初期化
	p := progress.New(progress.WithDefaultGradient())

	return &SettingsView{
		state:        state,
		styles:       styles,
		questionFlow: models.NewQuestionFlow(),
		textInput:    ti,
		progress:     p,
		selectIndex:  0,
	}
}

func (v *SettingsView) Init() tea.Cmd {
	// 最初の質問でTextInputを初期化
	v.updateTextInputForCurrentQuestion()
	return textinput.Blink
}

func (v *SettingsView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return v.handleKeyPress(msg)
	}

	// TextInputの更新
	var cmd tea.Cmd
	v.textInput, cmd = v.textInput.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *SettingsView) handleKeyPress(msg tea.KeyMsg) (ViewController, tea.Cmd) {
	currentQuestion := v.questionFlow.GetCurrentQuestion()
	if currentQuestion == nil {
		return v, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return v, tea.Quit

	case "esc":
		// 前の画面に戻る
		if v.questionFlow.CurrentIndex == 0 {
			return v, func() tea.Msg {
				return StepChangeMsg{Step: models.StepTemplateSelection}
			}
		}
		// 前の質問に戻る
		v.questionFlow.GoToPrevious()
		v.updateTextInputForCurrentQuestion()
		v.errorMessage = ""

	case "f1", "?":
		// ヘルプの表示切り替え
		v.showHelp = !v.showHelp

	case "enter":
		// 回答を設定
		var answer string

		switch currentQuestion.Type {
		case models.QuestionTypeText, models.QuestionTypeBool:
			answer = v.textInput.Value()
		case models.QuestionTypeSelect:
			if v.selectIndex < len(currentQuestion.Options) {
				answer = currentQuestion.Options[v.selectIndex]
			}
		}

		// バリデーション実行
		err := v.questionFlow.SetAnswer(answer)
		if err != nil {
			v.errorMessage = err.Error()
			return v, nil
		}

		// エラーメッセージをクリア
		v.errorMessage = ""

		// 次の質問に進む
		if !v.questionFlow.GoToNext() {
			// 全ての質問が完了
			v.completeSettings()
			return v, func() tea.Msg {
				return StepChangeMsg{Step: models.StepConfirmation}
			}
		}

		// 次の質問のためにTextInputを更新
		v.updateTextInputForCurrentQuestion()

	case "up", "k":
		// Select型質問での選択肢移動
		if currentQuestion.Type == models.QuestionTypeSelect {
			if v.selectIndex > 0 {
				v.selectIndex--
			}
		}

	case "down", "j":
		// Select型質問での選択肢移動
		if currentQuestion.Type == models.QuestionTypeSelect {
			if v.selectIndex < len(currentQuestion.Options)-1 {
				v.selectIndex++
			}
		}

	case "ctrl+u":
		// 入力をクリア
		v.textInput.SetValue("")

	default:
		// Select型以外の場合はTextInputに転送
		if currentQuestion.Type != models.QuestionTypeSelect {
			var cmd tea.Cmd
			v.textInput, cmd = v.textInput.Update(msg)
			return v, cmd
		}
	}

	return v, nil
}

func (v *SettingsView) updateTextInputForCurrentQuestion() {
	question := v.questionFlow.GetCurrentQuestion()
	if question == nil {
		return
	}

	// 既存の回答があれば設定
	if answer := v.questionFlow.GetAnswer(question.ID); answer != nil {
		v.textInput.SetValue(answer.Value)
	} else {
		v.textInput.SetValue(question.DefaultValue)
	}

	// TextInputの設定を質問タイプに応じて調整
	switch question.Type {
	case models.QuestionTypeText:
		v.textInput.Placeholder = "入力してください..."
		v.textInput.CharLimit = 100

	case models.QuestionTypeBool:
		v.textInput.Placeholder = "y/n"
		v.textInput.CharLimit = 10

	case models.QuestionTypeSelect:
		// Select型の場合、最初のオプションを選択
		v.selectIndex = 0
		if question.DefaultValue != "" {
			for i, option := range question.Options {
				if option == question.DefaultValue {
					v.selectIndex = i
					break
				}
			}
		}
	}
}

func (v *SettingsView) completeSettings() {
	// 質問フローからRepositoryConfigを生成
	repoConfig := v.questionFlow.ToRepositoryConfig()
	v.state.RepoConfig = repoConfig
}

func (v *SettingsView) View() string {
	if v.width == 0 {
		return "初期化中..."
	}

	question := v.questionFlow.GetCurrentQuestion()
	if question == nil {
		return "質問の読み込みに失敗しました"
	}

	var sections []string

	// プログレス表示
	current, total := v.questionFlow.GetProgress()
	progressRatio := v.questionFlow.GetProgressRatio()
	v.progress.Width = v.width - 4

	progressSection := lipgloss.JoinVertical(
		lipgloss.Left,
		v.styles.Title.Render(fmt.Sprintf("設定 (%d/%d)", current, total)),
		v.progress.ViewAs(progressRatio),
	)
	sections = append(sections, progressSection)

	// 質問タイトル
	titleSection := v.styles.Title.Render("🔧 " + question.Title)
	sections = append(sections, titleSection)

	// 質問の説明
	if question.Description != "" {
		descSection := v.styles.Text.Render(question.Description)
		sections = append(sections, descSection)
	}

	// 入力エリア
	inputSection := v.renderInputArea(question)
	sections = append(sections, inputSection)

	// エラーメッセージ
	if v.errorMessage != "" {
		errorSection := v.styles.Error.Render("❌ " + v.errorMessage)
		sections = append(sections, errorSection)
	}

	// ヘルプテキスト
	if question.HelpText != "" {
		helpIcon := "💡"
		if question.Required {
			helpIcon = "📝"
		}
		helpSection := v.styles.Debug.Render(helpIcon + " " + question.HelpText)
		sections = append(sections, helpSection)
	}

	// 拡張ヘルプ（F1で表示切り替え）
	if v.showHelp {
		helpSection := v.renderExtendedHelp()
		sections = append(sections, helpSection)
	}

	// キーバインドヘルプ
	keybindSection := v.renderKeybindHelp(question)
	sections = append(sections, keybindSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *SettingsView) renderInputArea(question *models.Question) string {
	switch question.Type {
	case models.QuestionTypeText, models.QuestionTypeBool:
		// テキスト入力
		return v.styles.Focused.Render(v.textInput.View())

	case models.QuestionTypeSelect:
		// 選択肢表示
		var options []string
		for i, option := range question.Options {
			if i == v.selectIndex {
				options = append(options, v.styles.Selected.Render("▶ "+option))
			} else {
				options = append(options, v.styles.Unselected.Render("  "+option))
			}
		}
		return v.styles.Border.Render(strings.Join(options, "\n"))

	default:
		return v.styles.Error.Render("未対応の質問タイプです")
	}
}

func (v *SettingsView) renderExtendedHelp() string {
	helpLines := []string{
		"📚 詳細ヘルプ:",
		"",
		"• Enter: 回答を確定して次へ進む",
		"• Esc: 前の質問に戻る（最初の質問では前の画面に戻る）",
		"• ↑↓/k/j: 選択肢を移動（選択式質問）",
		"• Ctrl+U: 入力をクリア",
		"• F1/?:ヘルプ表示を切り替え",
		"• Ctrl+C: アプリケーション終了",
		"",
		"💡 入力のコツ:",
		"• 必須項目は 📝 マークが付いています",
		"• 任意項目は空欄でもOKです",
		"• Yes/No質問は y/n で簡単に回答できます",
	}

	return v.styles.Info.Render(strings.Join(helpLines, "\n"))
}

func (v *SettingsView) renderKeybindHelp(question *models.Question) string {
	var keys []string

	switch question.Type {
	case models.QuestionTypeSelect:
		keys = append(keys, "↑↓: 選択")
	case models.QuestionTypeText:
		keys = append(keys, "テキスト入力")
	case models.QuestionTypeBool:
		keys = append(keys, "y/n: Yes/No")
	}

	keys = append(keys, "Enter: 決定", "Esc: 戻る", "F1: ヘルプ")

	return v.styles.Debug.Render("⌨️  " + strings.Join(keys, "  "))
}

func (v *SettingsView) SetSize(width, height int) {
	v.width = width
	v.height = height

	// TextInputの幅を調整
	v.textInput.Width = width - 10
	if v.textInput.Width < 20 {
		v.textInput.Width = 20
	}

	// Progressバーの幅を調整
	v.progress.Width = width - 4
}

func (v *SettingsView) GetTitle() string {
	return "リポジトリ設定"
}

func (v *SettingsView) CanGoBack() bool {
	// 質問フロー中は常にtrue（ESCで前の質問または前の画面に戻れる）
	return true
}

func (v *SettingsView) CanGoNext() bool {
	// 全ての質問が完了し、かつ全ての回答が有効な場合のみtrue
	if !v.questionFlow.IsCompleted {
		return false
	}

	for _, answer := range v.questionFlow.Answers {
		if !answer.IsValid {
			return false
		}
	}

	return true
}
