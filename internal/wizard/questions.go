package wizard

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// Answers はSurveyの回答を格納する構造体
type Answers struct {
	Template     string `survey:"template"`
	ProjectName  string `survey:"projectName"`
	Description  string `survey:"description"`
	CreateGitHub bool   `survey:"createGitHub"`
	IsPrivate    bool   `survey:"isPrivate"`
}

// SurveyExecutor はsurvey実行のインターフェース
type SurveyExecutor interface {
	Ask(questions []*survey.Question, response interface{}) error
}

// DefaultSurveyExecutor はデフォルトのsurvey実行器
type DefaultSurveyExecutor struct{}

func (d *DefaultSurveyExecutor) Ask(questions []*survey.Question, response interface{}) error {
	return survey.Ask(questions, response)
}

// QuestionFlow は質問フローを管理する構造体
type QuestionFlow struct {
	templates      []models.Template
	answers        *Answers
	surveyExecutor SurveyExecutor
}

// NewQuestionFlow は新しい質問フローを作成する
func NewQuestionFlow(templates []models.Template) *QuestionFlow {
	return &QuestionFlow{
		templates:      templates,
		answers:        &Answers{},
		surveyExecutor: &DefaultSurveyExecutor{},
	}
}

// formatTemplateOption はテンプレート選択肢の表示形式を作成する
func formatTemplateOption(template models.Template) string {
	stars := ""
	if template.Stars > 0 {
		stars = fmt.Sprintf(" (⭐ %d)", template.Stars)
	}

	language := ""
	if template.Language != "" {
		language = fmt.Sprintf(" [%s]", template.Language)
	}

	return fmt.Sprintf("%s%s%s", template.Name, stars, language)
}

// findSelectedTemplate は選択されたテンプレートを取得する
func (qf *QuestionFlow) findSelectedTemplate() *models.Template {
	if qf.answers.Template == "テンプレートなし" {
		return nil
	}

	for _, template := range qf.templates {
		if formatTemplateOption(template) == qf.answers.Template {
			return &template
		}
	}
	return nil
}

// GetProjectConfig は回答からProjectConfigを生成する
func (qf *QuestionFlow) GetProjectConfig() *models.ProjectConfig {
	template := qf.findSelectedTemplate()

	return &models.ProjectConfig{
		Name:         qf.answers.ProjectName,
		Description:  qf.answers.Description,
		Template:     template,
		CreateGitHub: qf.answers.CreateGitHub,
		IsPrivate:    qf.answers.IsPrivate,
		LocalPath:    "./" + qf.answers.ProjectName,
	}
}

// CreateQuestions はテンプレート情報を基に質問を生成する
func (qf *QuestionFlow) CreateQuestions() []*survey.Question {
	// テンプレート選択肢を生成
	templateOptions := make([]string, len(qf.templates)+1)
	for i, t := range qf.templates {
		templateOptions[i] = formatTemplateOption(t)
	}
	templateOptions[len(qf.templates)] = "テンプレートなし"

	questions := []*survey.Question{
		{
			Name: "template",
			Prompt: &survey.Select{
				Message: "テンプレートを選択してください:",
				Options: templateOptions,
				Description: func(value string, index int) string {
					return "プロジェクトのベースとなるテンプレートを選択"
				},
			},
			Validate: survey.Required,
		},
		{
			Name: "projectName",
			Prompt: &survey.Input{
				Message: "プロジェクト名:",
				Help:    "英数字、ハイフン、アンダースコアが使用できます",
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "説明 (オプション):",
				Help:    "プロジェクトの説明を入力してください",
			},
		},
		{
			Name: "createGitHub",
			Prompt: &survey.Confirm{
				Message: "GitHubにリポジトリを作成しますか？",
				Default: true,
				Help:    "Noの場合はローカルにのみプロジェクトが作成されます",
			},
		},
	}

	return questions
}

// CreateConditionalQuestions は条件付き質問を生成する
func (qf *QuestionFlow) CreateConditionalQuestions() []*survey.Question {
	var questions []*survey.Question

	// GitHubリポジトリ作成時のみ表示される質問
	if qf.answers.CreateGitHub {
		questions = append(questions, &survey.Question{
			Name: "isPrivate",
			Prompt: &survey.Confirm{
				Message: "プライベートリポジトリにしますか？",
				Default: true,
				Help:    "プライベート: あなたのみアクセス可能 / パブリック: 誰でもアクセス可能",
			},
		})
	}

	return questions
}

// Execute は質問フローを実行してProjectConfigを返す
func (qf *QuestionFlow) Execute() (*models.ProjectConfig, error) {
	// 基本的な質問を実行
	questions := qf.CreateQuestions()
	err := qf.surveyExecutor.Ask(questions, qf.answers)
	if err != nil {
		return nil, fmt.Errorf("基本質問の実行に失敗: %w", err)
	}

	// 条件付き質問を実行
	conditionalQuestions := qf.CreateConditionalQuestions()
	if len(conditionalQuestions) > 0 {
		err = qf.surveyExecutor.Ask(conditionalQuestions, qf.answers)
		if err != nil {
			return nil, fmt.Errorf("条件付き質問の実行に失敗: %w", err)
		}
	}

	// 回答をProjectConfigに変換
	config := &models.ProjectConfig{
		Name:         qf.answers.ProjectName,
		Description:  qf.answers.Description,
		CreateGitHub: qf.answers.CreateGitHub,
		IsPrivate:    qf.answers.IsPrivate,
	}

	// テンプレートを検索
	for _, template := range qf.templates {
		if formatTemplateOption(template) == qf.answers.Template {
			config.Template = &template
			break
		}
	}

	return config, nil
}
