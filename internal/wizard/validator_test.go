package wizard

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ValidatorTestSuite はバリデーションテストスイート
type ValidatorTestSuite struct {
	suite.Suite
	validator *ProjectNameValidator
}

func (suite *ValidatorTestSuite) SetupTest() {
	suite.validator = NewProjectNameValidator()
}

func TestValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

func (suite *ValidatorTestSuite) TestValidateBasicRules() {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
		errorMsg    string
	}{
		// 有効なケース
		{"有効な英数字", "my-awesome-project", false, ""},
		{"数字入り", "project123", false, ""},
		{"アンダースコア", "my_project", false, ""},
		{"ピリオド入り", "my.project", false, ""},
		{"短い名前", "a", false, ""},
		{"境界値_100文字", strings.Repeat("a", 100), false, ""},

		// 無効なケース - 基本ルール
		{"空文字", "", true, "required"},
		{"スペースのみ", "   ", true, "required"},
		{"長すぎる", strings.Repeat("a", 101), true, "at most 100 characters"},
		{"無効な文字_スペース", "my project", true, "invalid characters"},
		{"無効な文字_日本語", "プロジェクト", true, "invalid characters"},
		{"無効な文字_特殊記号", "project@#$", true, "invalid characters"},
		{"無効な文字_絵文字", "project🧙‍♂️", true, "invalid characters"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.validateBasicRules(tt.projectName)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errorMsg != "" {
					suite.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ValidatorTestSuite) TestValidateGitHubRules() {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
		errorMsg    string
	}{
		// 有効なケース
		{"通常の名前", "normal-project", false, ""},
		{"数字で開始", "123project", false, ""},
		{"文字で終了", "project123", false, ""},

		// 無効なケース - GitHub規則
		{"先頭ピリオド", ".project", true, "cannot start with a period"},
		{"末尾ピリオド", "project.", true, "cannot end with a period"},
		{"連続ピリオド", "my..project", true, "consecutive periods"},
		{"連続ハイフン", "my--project", true, "consecutive hyphens"},
		{"連続アンダースコア", "my__project", true, "consecutive underscores"},
		{"先頭ハイフン", "-project", true, "cannot start with a hyphen"},
		{"末尾ハイフン", "project-", true, "cannot end with a hyphen"},
		{"先頭アンダースコア", "_project", true, "cannot start with an underscore"},
		{"末尾アンダースコア", "project_", true, "cannot end with an underscore"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.validateGitHubRules(tt.projectName)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errorMsg != "" {
					suite.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ValidatorTestSuite) TestValidateReservedNames() {
	tests := []struct {
		name         string
		projectName  string
		wantErr      bool
		caseVariants bool
	}{
		// Windows予約名
		{"Windows予約名_CON", "CON", true, true},
		{"Windows予約名_PRN", "PRN", true, true},
		{"Windows予約名_AUX", "AUX", true, true},
		{"Windows予約名_NUL", "NUL", true, true},
		{"Windows予約名_COM1", "COM1", true, true},
		{"Windows予約名_LPT1", "LPT1", true, true},

		// Git/GitHub予約名
		{"Git予約名_.", ".", true, false},
		{"Git予約名_..", "..", true, false},
		{"Git予約名_.git", ".git", true, true},
		{"Git予約名_.github", ".github", true, true},

		// 一般的な予約名
		{"一般予約名_admin", "admin", true, true},
		{"一般予約名_root", "root", true, true},
		{"一般予約名_api", "api", true, true},

		// 有効な名前（予約名と似ているが有効）
		{"有効_connection", "connection", false, false},
		{"有効_admin-panel", "admin-panel", false, false},
		{"有効_my-api", "my-api", false, false},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.validateReservedNames(tt.projectName)

			if tt.wantErr {
				suite.Error(err)
				suite.Contains(err.Error(), "reserved name")
			} else {
				suite.NoError(err)
			}

			// 大文字小文字の variant もテスト
			if tt.caseVariants && tt.wantErr {
				upperErr := suite.validator.validateReservedNames(strings.ToUpper(tt.projectName))
				suite.Error(upperErr, "大文字版も予約名としてエラーになるはず")

				lowerErr := suite.validator.validateReservedNames(strings.ToLower(tt.projectName))
				suite.Error(lowerErr, "小文字版も予約名としてエラーになるはず")
			}
		})
	}
}

func (suite *ValidatorTestSuite) TestValidateAdvancedRules() {
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
		errorMsg    string
	}{
		// 有効なケース
		{"バランスの良い名前", "my-project-2023", false, ""},
		{"適度な特殊文字", "web.app", false, ""},
		{"数字と文字の混合", "project123", false, ""},

		// 無効なケース - 高度なルール
		{"全て数字", "12345", true, "all-numeric"},
		{"特殊文字過多", "a.b-c_d.e-f_g", true, "too many special characters"},
		{"制御文字", "project\n", true, "control characters"},
		{"制御文字_タブ", "project\t", true, "control characters"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.validateAdvancedRules(tt.projectName)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errorMsg != "" {
					suite.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ValidatorTestSuite) TestValidate_IntegrationTest() {
	// 全体のバリデーションの統合テスト
	tests := []struct {
		name        string
		projectName string
		wantErr     bool
	}{
		{"完全に有効", "my-awesome-project", false},
		{"境界値_最短", "a", false},
		{"境界値_最長", strings.Repeat("a", 100), false},
		{"複数エラー_空文字", "", true},
		{"複数エラー_予約名", "CON", true},
		{"複数エラー_無効文字", "project with spaces", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.Validate(tt.projectName)

			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// ヘルパー関数のテスト
func TestHelperFunctions(t *testing.T) {
	t.Run("isAllDigits", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"12345", true},
			{"0", true},
			{"123a", false},
			{"", false},
			{"a123", false},
		}

		for _, tt := range tests {
			result := isAllDigits(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("countSpecialChars", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int
		}{
			{"project", 0},
			{"my-project", 1},
			{"my_project.js", 2},
			{"a.b-c_d", 3},
			{"normal123", 0},
		}

		for _, tt := range tests {
			result := countSpecialChars(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})

	t.Run("containsControlChars", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"normal", false},
			{"with\nnewline", true},
			{"with\ttab", true},
			{"with\rcarriage", true},
			{"", false},
		}

		for _, tt := range tests {
			result := containsControlChars(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
		}
	})
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name        string
		description interface{}
		wantErr     bool
		errorMsg    string
	}{
		// 有効なケース
		{"空の説明", "", false, ""},
		{"通常の説明", "This is a test project", false, ""},
		{"日本語説明", "これはテストプロジェクトです", false, ""},
		{"絵文字入り", "Test project 🧙‍♂️", false, ""},
		{"境界値_500文字", strings.Repeat("a", 500), false, ""},
		{"改行入り_5行", "line1\nline2\nline3\nline4\nline5", false, ""},

		// 無効なケース
		{"非文字列型", 123, true, "invalid input type"},
		{"長すぎる", strings.Repeat("a", 501), true, "at most 500 characters"},
		{"改行過多", strings.Repeat("line\n", 6), true, "at most 5 lines"},
		{"制御文字", "description\x00", true, "control characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.description)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateValidator_ValidateTemplateSelection(t *testing.T) {
	templates := []models.Template{
		{
			Name:       "valid-template",
			FullName:   "user/valid-template",
			IsTemplate: true,
			Private:    false,
			Stars:      10,
			Language:   "Go",
		},
		{
			Name:       "private-template",
			FullName:   "user/private-template",
			IsTemplate: true,
			Private:    true,
		},
		{
			Name:       "not-template",
			FullName:   "user/not-template",
			IsTemplate: false,
			Private:    false,
		},
	}

	validator := NewTemplateValidator(templates)

	tests := []struct {
		name      string
		selection string
		wantErr   bool
		errorMsg  string
	}{
		// 有効なケース
		{"有効なテンプレート", "valid-template (⭐ 10) [Go]", false, ""},
		{"テンプレートなし", "No template", false, ""},

		// 無効なケース
		{"空の選択", "", true, "please select"},
		{"存在しないテンプレート", "nonexistent-template", true, "not available"},
		{"テンプレートでないリポジトリ", "not-template", true, "not configured as a template"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTemplateSelection(tt.selection)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateValidator_GetSurveyValidator(t *testing.T) {
	templates := []models.Template{
		{Name: "test", FullName: "user/test", IsTemplate: true},
	}

	validator := NewTemplateValidator(templates)
	surveyValidator := validator.GetSurveyValidator()

	// 関数が返されることを確認
	require.NotNil(t, surveyValidator)

	// 有効な選択のテスト
	err := surveyValidator("No template")
	assert.NoError(t, err)

	// 無効な型のテスト
	err = surveyValidator(123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input type")
}

func TestConfigValidator_ValidateProjectConfig(t *testing.T) {
	templates := []models.Template{
		{
			Name:       "valid-template",
			FullName:   "user/valid-template",
			IsTemplate: true,
			Stars:      5,
			Language:   "Go",
		},
	}

	validator := NewConfigValidator(templates)

	tests := []struct {
		name    string
		config  *models.ProjectConfig
		wantErr bool
		errType string
	}{
		{
			name: "有効な設定_テンプレートあり",
			config: &models.ProjectConfig{
				Name: "valid-project",
				Template: &models.Template{
					Name:       "valid-template",
					FullName:   "user/valid-template",
					IsTemplate: true,
				},
				CreateGitHub: true,
			},
			wantErr: false,
		},
		{
			name: "有効な設定_テンプレートなし",
			config: &models.ProjectConfig{
				Name:         "valid-project",
				Description:  "Valid description",
				Template:     nil,
				CreateGitHub: false,
			},
			wantErr: false,
		},
		{
			name: "無効_プロジェクト名",
			config: &models.ProjectConfig{
				Name: "",
			},
			wantErr: true,
			errType: "project name",
		},
		{
			name: "無効_テンプレート",
			config: &models.ProjectConfig{
				Name: "valid-project",
				Template: &models.Template{
					Name:       "invalid-template",
					IsTemplate: false,
				},
			},
			wantErr: true,
			errType: "template",
		},
		{
			name: "無効_ローカルパス",
			config: &models.ProjectConfig{
				Name:      "valid-project",
				LocalPath: "../invalid/path",
			},
			wantErr: true,
			errType: "relative path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProjectConfig(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != "" {
					assert.Contains(t, err.Error(), tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_ValidateGitHubConstraints(t *testing.T) {
	validator := NewConfigValidator([]models.Template{})

	tests := []struct {
		name    string
		config  *models.ProjectConfig
		wantErr bool
	}{
		{
			name: "GitHub作成なし",
			config: &models.ProjectConfig{
				CreateGitHub: false,
			},
			wantErr: false,
		},
		{
			name: "GitHub作成_有効な設定",
			config: &models.ProjectConfig{
				Name:         "valid-repo",
				CreateGitHub: true,
				IsPrivate:    true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateGitHubConstraints(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkProjectNameValidator_Validate(b *testing.B) {
	validator := NewProjectNameValidator()
	testName := "my-awesome-project-with-long-name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(testName)
	}
}

func BenchmarkProjectNameValidator_ValidateMany(b *testing.B) {
	validator := NewProjectNameValidator()
	testNames := []string{
		"project1",
		"my-awesome-project",
		"simple",
		"complex-project-name-2023",
		"test.project",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testNames {
			_ = validator.Validate(name)
		}
	}
}

func BenchmarkValidateDescription(b *testing.B) {
	description := "This is a test project description that contains multiple words and should be validated efficiently"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateDescription(description)
	}
}

// メモリ割り当てのベンチマーク
func BenchmarkProjectNameValidator_ValidateAllocs(b *testing.B) {
	validator := NewProjectNameValidator()
	testName := "my-awesome-project"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = validator.Validate(testName)
	}
}

// 正規表現のパフォーマンステスト
func BenchmarkRegexValidation(b *testing.B) {
	validator := NewProjectNameValidator()
	testName := "my-awesome-project-123"

	b.Run("ValidPattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validator.validPattern.MatchString(testName)
		}
	})

	b.Run("InvalidPatterns", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, pattern := range validator.invalidPatterns {
				_ = pattern.MatchString(testName)
			}
		}
	})
}

// 大量データでのテスト
func TestValidator_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("ストレステストをスキップ")
	}

	validator := NewProjectNameValidator()

	// 1000個の異なるプロジェクト名をテスト
	for i := 0; i < 1000; i++ {
		projectName := fmt.Sprintf("project-%d", i)
		err := validator.Validate(projectName)
		assert.NoError(t, err, "Project name should be valid: %s", projectName)
	}
}
