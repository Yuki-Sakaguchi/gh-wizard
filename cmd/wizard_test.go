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
	// Writerを閉じる
	suite.stdoutWriter.Close()
	suite.stderrWriter.Close()
	
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	
	// バッファサイズを制限してデッドロックを防ぐ
	go func() {
		io.Copy(stdoutBuf, suite.stdoutReader)
	}()
	go func() {
		io.Copy(stderrBuf, suite.stderrReader)
	}()
	
	// 少し待って出力を読み取る
	time.Sleep(100 * time.Millisecond)
	
	return stdoutBuf.String(), stderrBuf.String()
}

func TestWizardTestSuite(t *testing.T) {
	suite.Run(t, new(WizardTestSuite))
}

func (suite *WizardTestSuite) TestWizardCommand_Help() {
	// 実際のwizardCmdの基本チェック
	assert.Equal(suite.T(), "wizard", wizardCmd.Use)
	assert.Contains(suite.T(), wizardCmd.Long, "GitHub Repository Wizard")
	assert.Contains(suite.T(), wizardCmd.Short, "対話式リポジトリ作成ウィザード")
	
	// フラグが正しく定義されているかチェック
	templateFlag := wizardCmd.Flags().Lookup("template")
	require.NotNil(suite.T(), templateFlag)
	assert.Equal(suite.T(), "t", templateFlag.Shorthand)
	
	nameFlag := wizardCmd.Flags().Lookup("name")
	require.NotNil(suite.T(), nameFlag)
	assert.Equal(suite.T(), "n", nameFlag.Shorthand)
	
	dryRunFlag := wizardCmd.Flags().Lookup("dry-run")
	require.NotNil(suite.T(), dryRunFlag)
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
			
			// 標準エラーを文字列として直接キャプチャ
			var capturedOutput bytes.Buffer
			
			// 元のos.Stderrを保存
			oldStderr := os.Stderr
			defer func() { os.Stderr = oldStderr }()
			
			// パイプを作成
			r, w, _ := os.Pipe()
			os.Stderr = w
			
			// 別のgoroutineで出力を読み取り
			done := make(chan bool)
			go func() {
				io.Copy(&capturedOutput, r)
				done <- true
			}()

			// エラーハンドリング実行
			result := runner.handleError(tt.inputError)

			// パイプを閉じて出力完了を待つ
			w.Close()
			<-done
			os.Stderr = oldStderr

			assert.Error(t, result)
			
			output := capturedOutput.String()
			
			// 基本的なエラーメッセージの存在確認
			assert.Contains(t, output, "エラー:")
			
			if tt.expectRetry {
				assert.Contains(t, output, "しばらく待ってから再実行")
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

	// 標準出力をキャプチャ
	var capturedOutput bytes.Buffer
	
	// 元のos.Stdoutを保存
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	
	// パイプを作成
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// 別のgoroutineで出力を読み取り
	done := make(chan bool)
	go func() {
		io.Copy(&capturedOutput, r)
		done <- true
	}()

	// 設定表示実行
	runner.printConfiguration(config)

	// パイプを閉じて出力完了を待つ
	w.Close()
	<-done
	os.Stdout = oldStdout

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

// 注意: WizardRunner の実装は cmd/wizard.go に移動済み