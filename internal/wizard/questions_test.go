package wizard

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnswers_ToProjectConfig は回答からProjectConfigへの変換をテスト
func TestAnswers_ToProjectConfig(t *testing.T) {
	templates := []models.Template{
		{
			Name:     "test-template",
			FullName: "user/test-template",
			Stars:    10,
			Language: "Go",
		},
	}

	tests := []struct {
		name     string
		answers  Answers
		expected *models.ProjectConfig
		wantErr  bool
	}{
		{
			name: "valid answers with template",
			answers: Answers{
				Template:     "test-template (⭐ 10) [Go]",
				ProjectName:  "my-project",
				Description:  "Test project",
				CreateGitHub: true,
				IsPrivate:    true,
			},
			expected: &models.ProjectConfig{
				Name:         "my-project",
				Description:  "Test project",
				CreateGitHub: true,
				IsPrivate:    true,
				LocalPath:    "./my-project",
			},
			wantErr: false,
		},
		{
			name: "no template selected",
			answers: Answers{
				Template:     "No template",
				ProjectName:  "empty-project",
				Description:  "",
				CreateGitHub: false,
				IsPrivate:    false,
			},
			expected: &models.ProjectConfig{
				Name:         "empty-project",
				Description:  "",
				CreateGitHub: false,
				IsPrivate:    false,
				LocalPath:    "./empty-project",
				Template:     nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow := NewQuestionFlow(templates)
			flow.answers = &tt.answers

			config := flow.GetProjectConfig()

			require.NotNil(t, config)
			assert.Equal(t, tt.expected.Name, config.Name)
			assert.Equal(t, tt.expected.Description, config.Description)
			assert.Equal(t, tt.expected.CreateGitHub, config.CreateGitHub)
			assert.Equal(t, tt.expected.IsPrivate, config.IsPrivate)
			assert.Equal(t, tt.expected.LocalPath, config.LocalPath)
		})
	}
}

// TestQuestionFlow_CreateConditionalQuestions は条件付き質問生成のテスト
func TestQuestionFlow_CreateConditionalQuestions(t *testing.T) {
	flow := NewQuestionFlow([]models.Template{})

	tests := []struct {
		name              string
		createGitHub      bool
		expectedQuestions int
	}{
		{
			name:              "GitHub creation enabled",
			createGitHub:      true,
			expectedQuestions: 1, // IsPrivate question
		},
		{
			name:              "GitHub creation disabled",
			createGitHub:      false,
			expectedQuestions: 0, // No additional questions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flow.answers.CreateGitHub = tt.createGitHub

			questions := flow.CreateConditionalQuestions()
			assert.Len(t, questions, tt.expectedQuestions)

			if tt.expectedQuestions > 0 {
				// プライベートリポジトリの質問があることを確認
				assert.Equal(t, "isPrivate", questions[0].Name)
			}
		})
	}
}

func TestFormatTemplateOption(t *testing.T) {
	tests := []struct {
		name     string
		template models.Template
		expected string
	}{
		{
			name: "complete template info",
			template: models.Template{
				Name:     "nextjs-starter",
				Stars:    15,
				Language: "TypeScript",
			},
			expected: "nextjs-starter (⭐ 15) [TypeScript]",
		},
		{
			name: "no stars",
			template: models.Template{
				Name:     "simple-template",
				Stars:    0,
				Language: "JavaScript",
			},
			expected: "simple-template [JavaScript]",
		},
		{
			name: "no language",
			template: models.Template{
				Name:  "basic-template",
				Stars: 5,
			},
			expected: "basic-template (⭐ 5)",
		},
		{
			name: "minimal info",
			template: models.Template{
				Name: "minimal",
			},
			expected: "minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTemplateOption(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// MockSurveyExecutor はテスト用のモック
type MockSurveyExecutor struct {
	MockAnswers *Answers
	CallCount   int
}

func (m *MockSurveyExecutor) Ask(questions []*survey.Question, response interface{}) error {
	m.CallCount++
	if answers, ok := response.(*Answers); ok && m.MockAnswers != nil {
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
	assert.Equal(t, 3, mockExecutor.CallCount) // テンプレート選択 + 基本質問 + 条件付き質問
}

func TestFormatDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		maxLength   int
		expected    string
	}{
		{
			name:        "short description",
			description: "A simple template",
			maxLength:   50,
			expected:    "A simple template",
		},
		{
			name:        "long description needs truncation",
			description: "This is a very long description that exceeds the maximum character limit and should be truncated properly",
			maxLength:   50,
			expected:    "This is a very long description that exceeds th...",
		},
		{
			name:        "empty description",
			description: "",
			maxLength:   50,
			expected:    "No description available",
		},
		{
			name:        "exact length description",
			description: "Exactly fifty characters in this description!",
			maxLength:   50,
			expected:    "Exactly fifty characters in this description!",
		},
		{
			name:        "very short max length",
			description: "Short description",
			maxLength:   10,
			expected:    "Short d...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDescription(tt.description, tt.maxLength)
			assert.Equal(t, tt.expected, result)
			
			// Ensure result doesn't exceed maxLength
			assert.LessOrEqual(t, len(result), tt.maxLength)
		})
	}
}

func TestFormatDescriptionForTerminal(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    func(string) bool // Function to validate the result
	}{
		{
			name:        "empty description",
			description: "",
			expected: func(result string) bool {
				return result == "No description available"
			},
		},
		{
			name:        "short description",
			description: "Short desc",
			expected: func(result string) bool {
				return result == "Short desc"
			},
		},
		{
			name:        "japanese description",
			description: "これは日本語の説明文です。とても長い説明文になっています。",
			expected: func(result string) bool {
				// Should be truncated and end with "..."
				return len(result) < len("これは日本語の説明文です。とても長い説明文になっています。") && 
					   result[len(result)-3:] == "..."
			},
		},
		{
			name:        "mixed japanese and english",
			description: "This is a mixed 日本語 and English description that might be very long",
			expected: func(result string) bool {
				// Should handle mixed characters properly
				return len(result) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDescriptionForTerminal(tt.description)
			assert.True(t, tt.expected(result), "Result: %s", result)
		})
	}
}

func TestGetStringDisplayWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "ascii characters",
			input:    "Hello",
			expected: 5,
		},
		{
			name:     "japanese hiragana",
			input:    "こんにちは",
			expected: 10, // Each hiragana character takes 2 display columns
		},
		{
			name:     "japanese kanji",
			input:    "日本語",
			expected: 6, // Each kanji character takes 2 display columns
		},
		{
			name:     "mixed characters",
			input:    "Hello世界",
			expected: 9, // "Hello" (5) + "世界" (4) = 9
		},
		{
			name:     "emoji",
			input:    "👋🌍",
			expected: 4, // Each emoji takes 2 display columns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringDisplayWidth(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
