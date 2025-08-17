package wizard

import (
    "testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SurveyExecutor はSurvey実行を抽象化するインターフェース
type SurveyExecutor interface {
    Ask(questions []*survey.Question, response interface{}) error
}

// DefaultSurveyExecutor は実際のSurvey実行
type DefaultSurveyExecutor struct{}

func (e *DefaultSurveyExecutor) Ask(questions []*survey.Question, response interface{}) error {
    return survey.Ask(questions, response)
}

// MockSurveyExecutor はテスト用のモック
type MockSurveyExecutor struct {
    MockAnswers *Answers
    CallCount   int
}

func (m *MockSurveyExecutor) Ask(questions []*survey.Question, response interface{}) error {
    m.CallCount++
    if answers, ok := response.(*Answers); ok && m.MockAnswers \!= nil {
        *answers = *m.MockAnswers
    }
    return nil
}

// TestQuestionFlow_ExecuteWithMock はモックを使った実行テスト
func TestQuestionFlow_ExecuteWithMock(t *testing.T) {
    templates := []models.Template{
        {Name: "test-template", Stars: 5, Language: "Go"},
    }

    mockAnswers := &Answers{
        Template:     "test-template (⭐ 5) [Go]",
        ProjectName:  "test-project",
        Description:  "Test description",
        CreateGitHub: true,
        IsPrivate:    false,
    }

    mockExecutor := &MockSurveyExecutor{MockAnswers: mockAnswers}
    flow := NewQuestionFlow(templates)
    flow.surveyExecutor = mockExecutor

    config, err := flow.Execute()

    require.NoError(t, err)
    assert.Equal(t, "test-project", config.Name)
    assert.Equal(t, "Test description", config.Description)
    assert.True(t, config.CreateGitHub)
    assert.False(t, config.IsPrivate)
    assert.Equal(t, 2, mockExecutor.CallCount) // メイン質問 + 条件付き質問
}