package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/spf13/cobra"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/models"
)

// WizardTestSuite は統合テストスイート
type WizardTestSuite struct {
	suite.Suite
	originalStdout *os.File
	originalStderr *os.File
	stdoutReader   *os.File
	stdoutWriter   *os.File
	stderrReader   *os.File
	stderrWriter   *os.File
}

func (suite *WizardTestSuite) SetupTest() {
	// 標準出力をキャプチャするためのパイプを作成
	suite.originalStdout = os.Stdout
	suite.originalStderr = os.Stderr
	
	suite.stdoutReader, suite.stdoutWriter, _ = os.Pipe()
	suite.stderrReader, suite.stderrWriter, _ = os.Pipe()
	
	os.Stdout = suite.stdoutWriter
	os.Stderr = suite.stderrWriter
}

func (suite *WizardTestSuite) TearDownTest() {
	// 標準出力を復元
	os.Stdout = suite.originalStdout
	os.Stderr = suite.originalStderr
	
	suite.stdoutWriter.Close()
	suite.stderrWriter.Close()
	suite.stdoutReader.Close()
	suite.stderrReader.Close()
}

func (suite *WizardTestSuite) captureOutput() (string, string) {
	suite.stdoutWriter.Close()
	suite.stderrWriter.Close()
	
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	
	io.Copy(stdoutBuf, suite.stdoutReader)
	io.Copy(stderrBuf, suite.stderrReader)
	
	return stdoutBuf.String(), stderrBuf.String()
}

func TestWizardTestSuite(t *testing.T) {
	suite.Run(t, new(WizardTestSuite))
}

func (suite *WizardTestSuite) TestWizardCommand_Help() {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Test wizard command",
		RunE:  runWizard,
	}
	
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	
	require.NoError(suite.T(), err)
	
	stdout, _ := suite.captureOutput()
	assert.Contains(suite.T(), stdout, "wizard")
	assert.Contains(suite.T(), stdout, "GitHub Repository Wizard")
}

func TestWizardRunner_CheckPrerequisites(t *testing.T) {
	tests := []struct {
		name          string
		skipGitCheck  bool
		skipGHCheck   bool
		expectedError bool
	}{
		{
			name:          "all prerequisites available",
			skipGitCheck:  false,
			skipGHCheck:   false,
			expectedError: false,
		},
		{
			name:          "git not available",
			skipGitCheck:  true,
			skipGHCheck:   false,
			expectedError: true,
		},
		{
			name:          "gh not available",
			skipGitCheck:  false,
			skipGHCheck:   true,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipGitCheck || tt.skipGHCheck {
				t.Skip("環境依存のため統合テストでのみ実行")
			}

			runner := &WizardRunner{}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := runner.checkPrerequisites(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				// 開発環境では通常成功するはず
				assert.NoError(t, err)
			}
		})
	}
}

func TestWizardRunner_NonInteractiveMode(t *testing.T) {
	tests := []struct {
		name         string
		templateFlag string
		nameFlag     string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid flags",
			templateFlag: "user/template",
			nameFlag:     "test-project",
			expectError:  false,
		},
		{
			name:         "missing name flag",
			templateFlag: "user/template",
			nameFlag:     "",
			expectError:  true,
			errorMsg:     "--name フラグが必要",
		},
		{
			name:         "no template specified",
			templateFlag: "none",
			nameFlag:     "test-project",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &WizardRunner{}
			
			// モックテンプレートを用意
			templates := []models.Template{
				{
					Name:     "template",
					FullName: "user/template",
					Stars:    5,
				},
			}

			config, err := runner.runNonInteractiveMode(templates, tt.templateFlag, tt.nameFlag)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.nameFlag, config.Name)
				
				if tt.templateFlag == "none" || tt.templateFlag == "" {
					assert.Nil(t, config.Template)
				} else {
					assert.NotNil(t, config.Template)
					assert.Equal(t, tt.templateFlag, config.Template.FullName)
				}
			}
		})
	}
}

func TestWizardRunner_HandleError(t *testing.T) {
	tests := []struct {
		name        string
		inputError  error
		expectRetry bool
	}{
		{
			name:        "validation error",
			inputError:  models.NewValidationError("invalid input"),
			expectRetry: false,
		},
		{
			name:        "github error (retryable)",
			inputError:  models.NewGitHubError("api failed", nil),
			expectRetry: true,
		},
		{
			name:        "generic error",
			inputError:  errors.New("unknown error"),
			expectRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &WizardRunner{}
			
			// エラーハンドリングの結果をキャプチャ
			var capturedOutput bytes.Buffer
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w
			
			go func() {
				defer w.Close()
				io.Copy(&capturedOutput, r)
			}()

			result := runner.handleError(tt.inputError)

			os.Stderr = oldStderr
			w.Close()

			assert.Error(t, result)
			
			output := capturedOutput.String()
			
			if tt.expectRetry {
				assert.Contains(t, output, "しばらく待ってから再実行")
			}
			
			// エラーメッセージが適切にフォーマットされているかチェック
			if wizardErr, ok := tt.inputError.(*models.WizardError); ok {
				assert.Contains(t, output, wizardErr.Message)
			}
		})
	}
}

func TestWizardRunner_PrintConfiguration(t *testing.T) {
	config := &models.ProjectConfig{
		Name:         "test-project",
		Description:  "Test description",
		CreateGitHub: true,
		IsPrivate:    true,
		LocalPath:    "./test-project",
	}

	runner := &WizardRunner{}

	// 出力をキャプチャ
	var capturedOutput bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	go func() {
		defer w.Close()
		io.Copy(&capturedOutput, r)
	}()

	runner.printConfiguration(config)

	os.Stdout = oldStdout
	w.Close()

	output := capturedOutput.String()

	// 設定内容が正しく表示されているかチェック
	assert.Contains(t, output, "設定内容確認")
	assert.Contains(t, output, "test-project")
	assert.Contains(t, output, "Test description")
	assert.Contains(t, output, "プライベート")
}

func TestWizardRunner_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("パフォーマンステストをスキップ")
	}

	runner := &WizardRunner{}
	
	// モックデータの準備
	templates := make([]models.Template, 100) // 大量のテンプレート
	for i := 0; i < 100; i++ {
		templates[i] = models.Template{
			Name:     fmt.Sprintf("template-%d", i),
			FullName: fmt.Sprintf("user/template-%d", i),
			Stars:    i,
		}
	}

	start := time.Now()
	
	config, err := runner.runNonInteractiveMode(templates, "none", "perf-test")
	
	elapsed := time.Since(start)
	
	require.NoError(t, err)
	assert.NotNil(t, config)
	
	// パフォーマンス要件: 1秒以内で完了
	assert.Less(t, elapsed, time.Second, "Non-interactive mode should complete within 1 second")
}

func BenchmarkWizardRunner_NonInteractiveMode(b *testing.B) {
	runner := &WizardRunner{}
	templates := []models.Template{
		{Name: "template", FullName: "user/template", Stars: 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.runNonInteractiveMode(templates, "none", "bench-test")
	}
}

// WizardRunner は未実装のため、テスト用のスタブを定義
type WizardRunner struct{}

func (wr *WizardRunner) checkPrerequisites(ctx context.Context) error {
	// TODO: 実装予定
	return nil
}

func (wr *WizardRunner) runNonInteractiveMode(templates []models.Template, templateFlag, nameFlag string) (*models.ProjectConfig, error) {
	// TODO: 実装予定
	if nameFlag == "" {
		return nil, errors.New("--name フラグが必要です")
	}
	
	return &models.ProjectConfig{
		Name: nameFlag,
	}, nil
}

func (wr *WizardRunner) handleError(err error) error {
	// TODO: 実装予定
	if wizardErr, ok := err.(*models.WizardError); ok && wizardErr.IsRetryable() {
		fmt.Fprintf(os.Stderr, "エラー: %s\nしばらく待ってから再実行してください\n", err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "エラー: %s\n", err.Error())
	}
	return err
}

func (wr *WizardRunner) printConfiguration(config *models.ProjectConfig) {
	// TODO: 実装予定
	fmt.Println("📋 設定内容確認")
	fmt.Printf("プロジェクト名: %s\n", config.Name)
	fmt.Printf("説明: %s\n", config.Description)
	if config.IsPrivate {
		fmt.Println("可視性: プライベート")
	} else {
		fmt.Println("可視性: パブリック")
	}
}