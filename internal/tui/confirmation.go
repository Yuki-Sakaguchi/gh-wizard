package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// ConfirmationView は確認画面
type ConfirmationView struct {
	state  *models.WizardState
	styles *Styles
	width  int
	height int

	// 確認画面のデータ
	confirmationData *models.ConfirmationData
	
	// UI状態
	selectedAction    int  // 選択中のアクション
	showWarnings     bool // 警告表示の切り替え
	showCommand      bool // 実行コマンド表示の切り替え
	
	// レイアウト設定
	maxSectionWidth  int
	contentPadding   int
}

func NewConfirmationView(state *models.WizardState, styles *Styles) *ConfirmationView {
	return &ConfirmationView{
		state:           state,
		styles:          styles,
		selectedAction:  2, // デフォルトは "リポジトリ作成"
		showWarnings:    false,
		showCommand:     false,
		contentPadding:  2,
	}
}

func (v *ConfirmationView) Init() tea.Cmd {
	// 確認データを構築
	v.confirmationData = models.BuildConfirmationData(v.state)
	return nil
}

func (v *ConfirmationView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return v.handleKeyPress(msg)
	}
	
	return v, nil
}

func (v *ConfirmationView) handleKeyPress(msg tea.KeyMsg) (ViewController, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return v, tea.Quit

	case "esc":
		// 前の画面（設定画面）に戻る
		return v, func() tea.Msg {
			return StepChangeMsg{Step: models.StepRepositorySettings}
		}

	case "left", "h":
		// アクション選択: 左へ
		if v.selectedAction > 0 {
			v.selectedAction--
		}

	case "right", "l":
		// アクション選択: 右へ
		if v.selectedAction < len(v.confirmationData.Actions)-1 {
			v.selectedAction++
		}

	case "1":
		// ショートカット: 設定修正
		v.selectedAction = 0
		return v.executeAction()

	case "2":
		// ショートカット: キャンセル
		v.selectedAction = 1
		return v.executeAction()

	case "3":
		// ショートカット: リポジトリ作成
		v.selectedAction = 2
		return v.executeAction()

	case "w", "W":
		// 警告表示の切り替え
		v.showWarnings = !v.showWarnings

	case "c", "C":
		// コマンド表示の切り替え
		v.showCommand = !v.showCommand

	case "enter":
		// 選択されたアクションを実行
		return v.executeAction()

	case "r", "R":
		// データを再構築（リフレッシュ）
		v.confirmationData = models.BuildConfirmationData(v.state)
	}

	return v, nil
}

func (v *ConfirmationView) executeAction() (ViewController, tea.Cmd) {
	if v.selectedAction < 0 || v.selectedAction >= len(v.confirmationData.Actions) {
		return v, nil
	}

	action := v.confirmationData.Actions[v.selectedAction]

	switch action {
	case models.ActionModifySettings:
		// リポジトリ設定画面に戻る
		return v, func() tea.Msg {
			return StepChangeMsg{Step: models.StepRepositorySettings}
		}

	case models.ActionCancel:
		// ウェルカム画面に戻る
		return v, func() tea.Msg {
			return StepChangeMsg{Step: models.StepWelcome}
		}

	case models.ActionCreateRepository:
		// 実行画面に進む
		return v, func() tea.Msg {
			return StepChangeMsg{Step: models.StepExecution}
		}

	default:
		return v, nil
	}
}

func (v *ConfirmationView) View() string {
	if v.width == 0 || v.confirmationData == nil {
		return "初期化中..."
	}

	v.calculateLayout()

	var sections []string

	// タイトル
	title := v.styles.Title.Render("📋 リポジトリ作成の確認")
	sections = append(sections, title)

	// 各セクションを表示
	for _, section := range v.confirmationData.Sections {
		sectionView := v.renderSection(section)
		sections = append(sections, sectionView)
	}

	// 警告表示（切り替え可能）
	if v.showWarnings && len(v.confirmationData.Warnings) > 0 {
		warningSection := v.renderWarnings()
		sections = append(sections, warningSection)
	}

	// コマンド表示（切り替え可能）
	if v.showCommand {
		commandSection := v.renderCommand()
		sections = append(sections, commandSection)
	}

	// アクションボタン
	actionsSection := v.renderActions()
	sections = append(sections, actionsSection)

	// ヘルプ
	helpSection := v.renderHelp()
	sections = append(sections, helpSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *ConfirmationView) calculateLayout() {
	// セクションの最大幅を計算
	availableWidth := v.width - (v.contentPadding * 2)
	v.maxSectionWidth = availableWidth
	if v.maxSectionWidth > 80 {
		v.maxSectionWidth = 80
	}
}

func (v *ConfirmationView) renderSection(section models.ConfirmationSection) string {
	var lines []string
	
	// セクションタイトル
	titleStyle := v.styles.Subtitle.Copy().
		Bold(true).
		Foreground(lipgloss.Color(v.styles.Colors.Primary))
	
	sectionTitle := titleStyle.Render(section.Icon + " " + section.Title)
	lines = append(lines, sectionTitle)

	// セクション内の項目
	for _, item := range section.Items {
		itemLine := v.renderItem(item)
		lines = append(lines, itemLine)
	}

	// セクション警告
	if section.HasWarning && section.Warning != "" {
		warningLine := v.styles.Warning.Render("⚠️  " + section.Warning)
		lines = append(lines, warningLine)
	}

	// セクションをボーダーで囲む
	content := strings.Join(lines, "\n")
	
	sectionStyle := v.styles.Border.Copy().
		Width(v.maxSectionWidth).
		Padding(1, 2)
	
	return sectionStyle.Render(content)
}

func (v *ConfirmationView) renderItem(item models.ConfirmationItem) string {
	// 値部分のスタイル
	valueStyle := v.styles.Text.Copy()
	if item.Important {
		valueStyle = valueStyle.Foreground(lipgloss.Color(v.styles.Colors.Primary))
	}
	if item.Warning {
		valueStyle = valueStyle.Foreground(lipgloss.Color(v.styles.Colors.Warning))
	}
	
	value := valueStyle.Render(item.Value)

	// ラベル幅を調整してアライメント
	labelWidth := 16
	if len(item.Label) > labelWidth {
		labelWidth = len(item.Label) + 2
	}
	
	// ラベル部分のスタイル
	labelStyle := v.styles.Text.Copy()
	if item.Important {
		labelStyle = labelStyle.Bold(true)
	}
	
	alignedLabel := fmt.Sprintf("%-*s", labelWidth, item.Label+":")
	styledLabel := labelStyle.Render(alignedLabel)
	
	line := styledLabel + " " + value

	// 説明がある場合は追加
	if item.Description != "" {
		descStyle := v.styles.Debug.Copy().Italic(true)
		descLine := descStyle.Render(fmt.Sprintf("%*s%s", labelWidth+1, "", item.Description))
		line += "\n" + descLine
	}

	return line
}

func (v *ConfirmationView) renderWarnings() string {
	if len(v.confirmationData.Warnings) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, v.styles.Warning.Render("⚠️  警告事項"))
	
	for i, warning := range v.confirmationData.Warnings {
		warningText := fmt.Sprintf("%d. %s", i+1, warning)
		lines = append(lines, v.styles.Warning.Render(warningText))
	}

	content := strings.Join(lines, "\n")
	
	warningStyle := v.styles.Border.Copy().
		Width(v.maxSectionWidth).
		Padding(1, 2).
		BorderForeground(lipgloss.Color(v.styles.Colors.Warning))
	
	return warningStyle.Render(content)
}

func (v *ConfirmationView) renderCommand() string {
	command := v.confirmationData.FormatRepositoryCommand(v.state)
	if len(command) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, v.styles.Info.Render("🔧 実行コマンド"))
	
	commandLine := "gh " + strings.Join(command, " ")
	lines = append(lines, v.styles.Debug.Render(commandLine))

	content := strings.Join(lines, "\n")
	
	commandStyle := v.styles.Border.Copy().
		Width(v.maxSectionWidth).
		Padding(1, 2).
		BorderForeground(lipgloss.Color(v.styles.Colors.Info))
	
	return commandStyle.Render(content)
}

func (v *ConfirmationView) renderActions() string {
	var actionButtons []string

	for i, action := range v.confirmationData.Actions {
		buttonText := fmt.Sprintf("[%s] %s", action.GetKey(), action.String())
		
		var buttonStyle lipgloss.Style
		if i == v.selectedAction {
			// 選択中のボタン
			buttonStyle = v.styles.Selected.Copy().
				Padding(0, 2).
				Margin(0, 1)
		} else {
			// 非選択のボタン
			buttonStyle = v.styles.Unselected.Copy().
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(v.styles.Colors.Debug))
		}
		
		actionButtons = append(actionButtons, buttonStyle.Render(buttonText))
	}

	buttonsLine := lipgloss.JoinHorizontal(lipgloss.Left, actionButtons...)
	
	return v.styles.Text.Render("実行するアクションを選択してください:") + "\n" + buttonsLine
}

func (v *ConfirmationView) renderHelp() string {
	var helpItems []string

	helpItems = append(helpItems, "←→/h/l: アクション選択")
	helpItems = append(helpItems, "1-3: ダイレクトアクション")
	helpItems = append(helpItems, "Enter: 実行")
	helpItems = append(helpItems, "W: 警告表示切り替え")
	helpItems = append(helpItems, "C: コマンド表示切り替え")
	helpItems = append(helpItems, "Esc: 戻る")

	// 現在の状態を表示
	if len(v.confirmationData.Warnings) > 0 {
		if v.showWarnings {
			helpItems = append(helpItems, "⚠️  警告表示中")
		} else {
			helpItems = append(helpItems, "⚠️  警告あり（Wで表示）")
		}
	}

	if v.showCommand {
		helpItems = append(helpItems, "🔧 コマンド表示中")
	}

	return v.styles.Debug.Render("⌨️  " + strings.Join(helpItems, "  "))
}

func (v *ConfirmationView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *ConfirmationView) GetTitle() string {
	return "確認"
}

func (v *ConfirmationView) CanGoBack() bool {
	return true
}

func (v *ConfirmationView) CanGoNext() bool {
	// リポジトリ設定が有効な場合のみ次に進める
	return v.state.RepoConfig != nil && v.state.RepoConfig.Validate() == nil
}
