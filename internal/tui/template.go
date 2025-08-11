package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// templateItem はリストアイテムのインターフェースを実装
type templateItem struct {
	template models.Template
}

func (t templateItem) FilterValue() string {
	return t.template.Name
}

func (t templateItem) Title() string {
	return t.template.Name
}

func (t templateItem) Description() string {
	if t.template.Language != "" {
		return fmt.Sprintf("%s - ⭐ %d", t.template.Language, t.template.Stars)
	}
	return fmt.Sprintf("⭐ %d", t.template.Stars)
}

// templateDelegate はリストアイテムの描画をカスタマイズ
type templateDelegate struct{
	styles *Styles
}

func (d templateDelegate) Height() int {
	return 1
}

func (d templateDelegate) Spacing() int {
	return 0
}

func (d templateDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d templateDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	t, ok := listItem.(templateItem)
	if !ok {
		return
	}

	var str string
	if index == m.Index() {
		// 選択中の項目
		str = d.styles.Selected.Render(fmt.Sprintf("▸ %s", t.Title()))
	} else {
		// 非選択の項目
		str = d.styles.Unselected.Render(fmt.Sprintf("  %s", t.Title()))
	}

	fmt.Fprint(w, str)
}

// FileNode はファイル構造を表現する
type FileNode struct {
	Name     string
	Type     string // "file" or "dir"
	Children []FileNode
}

// TemplateDetails はテンプレートの詳細情報
type TemplateDetails struct {
	*models.Template
	ReadmeContent string
	FileStructure []FileNode
}

// TemplateView はテンプレート選択画面
type TemplateView struct {
	state        *models.WizardState
	styles       *Styles
	githubClient github.Client
	width        int
	height       int

	// List Model
	list         list.Model
	loading      bool
	error        error

	// 詳細パネル
	detailsCache map[string]*TemplateDetails

	// レイアウト
	splitRatio   float64 // 左右の分割比率（デフォルト: 0.4）
}

func NewTemplateView(state *models.WizardState, styles *Styles, githubClient github.Client) *TemplateView {
	// リストモデルの初期化
	l := list.New([]list.Item{}, templateDelegate{styles: styles}, 0, 0)
	l.Title = "📚 テンプレート"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Debug
	l.Styles.HelpStyle = styles.Debug

	return &TemplateView{
		state:        state,
		styles:       styles,
		githubClient: githubClient,
		list:         l,
		loading:      true,
		detailsCache: make(map[string]*TemplateDetails),
		splitRatio:   0.4,
	}
}

func (v *TemplateView) Init() tea.Cmd {
	return tea.Batch(
		v.fetchTemplatesCmd(),
		v.list.StartSpinner(),
	)
}

func (v *TemplateView) Update(msg tea.Msg) (ViewController, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TemplatesLoadedMsg:
		v.loading = false
		v.list.StopSpinner()
		
		if msg.Error != nil {
			v.error = msg.Error
			return v, nil
		}

		// テンプレートをリストアイテムに変換
		items := make([]list.Item, len(msg.Templates))
		for i, template := range msg.Templates {
			items[i] = templateItem{template: template}
		}

		cmds = append(cmds, v.list.SetItems(items))
		return v, tea.Batch(cmds...)

	case tea.KeyMsg:
		if v.loading {
			return v, nil // ローディング中は操作を無視
		}

		if v.error != nil {
			switch msg.String() {
			case "r", "R":
				// リトライ
				v.error = nil
				v.loading = true
				return v, tea.Batch(
					v.fetchTemplatesCmd(),
					v.list.StartSpinner(),
				)
			}
			return v, nil
		}

		switch msg.String() {
		case "enter":
			// 選択されたテンプレートを取得
			if selectedItem, ok := v.list.SelectedItem().(templateItem); ok {
				v.state.SelectedTemplate = &selectedItem.template
				return v, func() tea.Msg {
					return StepChangeMsg{Step: models.StepRepositorySettings}
				}
			}
			return v, nil
		}
	}

	// リストモデルの更新
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func (v *TemplateView) View() string {
	if v.width == 0 {
		return "初期化中..."
	}

	leftWidth, rightWidth, contentHeight := v.calculateLayout()

	// 左側: テンプレートリスト
	v.list.SetWidth(leftWidth)
	v.list.SetHeight(contentHeight)
	leftPanel := v.list.View()

	// 右側: 詳細パネル
	var rightPanel string
	if v.loading {
		rightPanel = v.styles.Info.Render("読み込み中...")
	} else if v.error != nil {
		errorMsg := v.styles.Error.Render(fmt.Sprintf("エラー: %v", v.error))
		retryMsg := v.styles.Info.Render("\nR: リトライ  Esc: 戻る")
		rightPanel = errorMsg + retryMsg
	} else {
		selectedTemplate := v.getSelectedTemplate()
		rightPanel = v.renderDetailPanel(selectedTemplate)
	}

	// 右パネルの高さを調整
	rightPanelStyled := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight).
		Padding(1).
		Render(rightPanel)

	// 左右を結合
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		lipgloss.NewStyle().
			Width(1).
			Height(contentHeight).
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color(v.styles.Colors.Primary)).
			Render(""),
		rightPanelStyled,
	)

	// ヘルプを追加
	help := v.styles.Debug.Render("⌨️  ↑↓: 選択  Enter: 決定  /: 検索  Esc: 戻る")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		help,
	)
}

func (v *TemplateView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *TemplateView) GetTitle() string {
	return "テンプレート選択"
}

func (v *TemplateView) CanGoBack() bool {
	return !v.loading
}

func (v *TemplateView) CanGoNext() bool {
	return !v.loading && v.error == nil && len(v.list.Items()) > 0
}

// レイアウト計算
func (v *TemplateView) calculateLayout() (leftWidth, rightWidth, contentHeight int) {
	// ボーダーとパディングを考慮
	availableWidth := v.width - 4  // 左右のボーダー
	contentHeight = v.height - 4   // 上下のボーダー + ヘルプテキスト

	leftWidth = int(float64(availableWidth) * v.splitRatio)
	rightWidth = availableWidth - leftWidth - 1 // 分割線のスペース

	return leftWidth, rightWidth, contentHeight
}

// 選択中のテンプレートを取得
func (v *TemplateView) getSelectedTemplate() *models.Template {
	if selectedItem, ok := v.list.SelectedItem().(templateItem); ok {
		return &selectedItem.template
	}
	return nil
}

// 詳細パネルの描画
func (v *TemplateView) renderDetailPanel(template *models.Template) string {
	if template == nil {
		return v.styles.Debug.Render("テンプレートを選択してください")
	}

	var sections []string

	// タイトル
	title := v.styles.Title.Render("📋 " + template.Name)
	sections = append(sections, title)

	// 基本情報
	basicInfo := []string{
		fmt.Sprintf("作成者: %s", template.Owner),
		fmt.Sprintf("言語: %s", template.Language),
		fmt.Sprintf("スター: ⭐ %d", template.Stars),
	}
	if !template.UpdatedAt.IsZero() {
		basicInfo = append(basicInfo, fmt.Sprintf("最終更新: %s", template.UpdatedAt.Format("2006-01-02")))
	}

	basicSection := v.styles.Text.Render(strings.Join(basicInfo, "\n"))
	sections = append(sections, basicSection)

	// 説明
	if template.Description != "" {
		descSection := v.styles.Info.Render("説明:\n" + template.Description)
		sections = append(sections, descSection)
	}

	// トピック/タグ
	if len(template.Topics) > 0 {
		topics := "🏷️ タグ: " + strings.Join(template.Topics, ", ")
		topicsSection := v.styles.Debug.Render(topics)
		sections = append(sections, topicsSection)
	}

	// ファイル構造（詳細取得済みの場合）
	if details := v.getTemplateDetails(template.FullName); details != nil {
		filesSection := v.renderFileStructure(details.FileStructure)
		if filesSection != "" {
			sections = append(sections, filesSection)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// ファイル構造の描画
func (v *TemplateView) renderFileStructure(files []FileNode) string {
	if len(files) == 0 {
		return ""
	}

	title := v.styles.Subtitle.Render("📁 主要ファイル:")

	var fileLines []string
	maxFiles := 8
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	for i, file := range files {
		prefix := "├── "
		if i == len(files)-1 {
			prefix = "└── "
		}

		icon := "📄"
		if file.Type == "dir" {
			icon = "📁"
		}

		fileLines = append(fileLines, prefix+icon+" "+file.Name)
	}

	if len(v.getTemplateDetails("").FileStructure) > maxFiles {
		fileLines = append(fileLines, fmt.Sprintf("└── ... (他 %d ファイル)", len(v.getTemplateDetails("").FileStructure)-maxFiles))
	}

	fileList := v.styles.Debug.Render(strings.Join(fileLines, "\n"))

	return title + "\n" + fileList
}

// テンプレート詳細を取得（キャッシュから）
func (v *TemplateView) getTemplateDetails(fullName string) *TemplateDetails {
	if details, exists := v.detailsCache[fullName]; exists {
		return details
	}
	return nil
}

// テンプレート取得コマンド
func (v *TemplateView) fetchTemplatesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		templates, err := v.githubClient.GetTemplateRepositories(ctx)
		return TemplatesLoadedMsg{
			Templates: templates,
			Error:     err,
		}
	}
}

// テンプレート関連のメッセージ型
type TemplatesLoadedMsg struct {
	Templates []models.Template
	Error     error
}
