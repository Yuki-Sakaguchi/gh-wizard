package wizard

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ExecutorTestSuite はテストスイートを定義
type ExecutorTestSuite struct {
	suite.Suite
	tempDir    string
	executor   *ProjectExecutor
	mockClient *github.SimpleMockClient
}

func (suite *ExecutorTestSuite) SetupTest() {
	suite.tempDir = suite.T().TempDir()
	suite.mockClient = github.NewSimpleMockClient()
	suite.executor = NewProjectExecutor(suite.mockClient)
}

func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

func (suite *ExecutorTestSuite) TestCreateEmptyProject() {
	projectPath := filepath.Join(suite.tempDir, "test-project")

	config := &models.ProjectConfig{
		Name:         "test-project",
		Description:  "Test project",
		LocalPath:    projectPath,
		CreateGitHub: false, // ローカルのみ
	}

	err := suite.executor.createLocalDirectory(context.Background(), config)
	require.NoError(suite.T(), err)

	// ディレクトリが作成されているかチェック
	assert.DirExists(suite.T(), projectPath)

	// README.md が作成されているかチェック
	readmePath := filepath.Join(projectPath, "README.md")
	assert.FileExists(suite.T(), readmePath)

	// README.md の内容をチェック
	content, err := os.ReadFile(readmePath)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(content), "test-project")
	assert.Contains(suite.T(), string(content), "Test project")
}

func (suite *ExecutorTestSuite) TestCreateProjectWithTemplate() {
	// 模擬テンプレートディレクトリを作成
	templateDir := filepath.Join(suite.tempDir, "template")
	require.NoError(suite.T(), os.MkdirAll(templateDir, 0755))

	// テンプレートファイルを作成
	templateFiles := map[string]string{
		"package.json": `{"name": "template", "version": "1.0.0"}`,
		"src/main.js":  `console.log("Hello from template");`,
		"README.md":    "# Template Project",
	}

	for filePath, content := range templateFiles {
		fullPath := filepath.Join(templateDir, filePath)
		require.NoError(suite.T(), os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(suite.T(), os.WriteFile(fullPath, []byte(content), 0644))
	}

	projectPath := filepath.Join(suite.tempDir, "from-template")

	config := &models.ProjectConfig{
		Name:      "from-template",
		LocalPath: projectPath,
		Template: &models.Template{
			FullName: "test/template",
			CloneURL: templateDir, // テスト用にローカルパス使用
		},
		CreateGitHub: false,
	}

	err := suite.executor.createLocalDirectory(context.Background(), config)
	require.NoError(suite.T(), err)

	// プロジェクトディレクトリが作成されているかチェック
	assert.DirExists(suite.T(), projectPath)

	// テンプレートファイルがコピーされているかチェック
	for filePath := range templateFiles {
		copiedFile := filepath.Join(projectPath, filePath)
		assert.FileExists(suite.T(), copiedFile, "Template file should be copied: %s", filePath)
	}
}

func (suite *ExecutorTestSuite) TestCreateProject_DirectoryExists() {
	projectPath := filepath.Join(suite.tempDir, "existing-project")

	// 既存ディレクトリを作成
	require.NoError(suite.T(), os.MkdirAll(projectPath, 0755))

	config := &models.ProjectConfig{
		Name:      "existing-project",
		LocalPath: projectPath,
	}

	err := suite.executor.createLocalDirectory(context.Background(), config)

	// ディレクトリが既に存在する場合はエラーになるはず
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "既に存在")
}

func TestProjectExecutor_Execute_LocalOnly(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test-local-project")

	config := &models.ProjectConfig{
		Name:         "test-local-project",
		Description:  "Test local project",
		LocalPath:    projectPath,
		CreateGitHub: false, // ローカルのみ
	}

	mockClient := github.NewSimpleMockClient()
	executor := NewProjectExecutor(mockClient)

	err := executor.Execute(context.Background(), config)
	require.NoError(t, err)

	// プロジェクトディレクトリが作成されているかチェック
	assert.DirExists(t, projectPath)

	// .git ディレクトリが作成されているかチェック（Git初期化）
	gitDir := filepath.Join(projectPath, ".git")
	assert.DirExists(t, gitDir)
}
