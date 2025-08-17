package wizard

import (
	"testing"

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
				Template:     "テンプレートなし",
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
