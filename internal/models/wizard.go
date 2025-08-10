package models

// Step はウィザードのステップを表す
type Step int

const (
	StepWelcome Step = iota
	StepTemplateSelection
	StepRepositorySettings
	StepConfirmation
	StepExecution
	StepCompleted
)

// String はステップの文字列表現を返す
func (s Step) String() string {
	switch s {
	case StepWelcome:
		return "ようこそ"
	case StepTemplateSelection:
		return "テンプレート選択"
	case StepRepositorySettings:
		return "リポジトの設定"
	case StepConfirmation:
		return "確認"
	case StepExecution:
		return "実行"
	case StepCompleted:
		return "完了"
	default:
		return "未知のステップ"
	}
}

// WizardState はウィザード全体の状態を管理する
type WizardState struct {
	CurrentStep      Step
	UseTemplate      bool
	SelectedTemplate *Template
	RepoConfig       *RepositoryConfig
	SearchQuery      string
	Templates        []Template
}

// NewWizardState は新しいウィザード状態を作成する
func NewWizardState() *WizardState {
	return &WizardState{
		CurrentStep: StepWelcome,
		UseTemplate: false,
		RepoConfig: &RepositoryConfig{
			IsPrivate:  true,
			SholdClone: true,
			AddReadme:  true,
		},
		Templates: make([]Template, 0),
	}
}

// CanProceedToNext は次のステップに進めるかを判定する
func (ws *WizardState) CanProceedToNext() bool {
	switch ws.CurrentStep {
	case StepWelcome:
		return true
	case StepTemplateSelection:
		return !ws.UseTemplate || ws.SelectedTemplate != nil
	case StepRepositorySettings:
		return ws.RepoConfig != nil && ws.RepoConfig.Validate() == nil
	case StepConfirmation:
		return true
	default:
		return false
	}
}
