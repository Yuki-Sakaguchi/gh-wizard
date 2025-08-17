package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ProjectConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: ProjectConfig{
				Name:        "test-project",
				Description: "テストプロジェクト",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: ProjectConfig{
				Name: "",
			},
			wantErr: true,
			errMsg:  "プロジェクト名は必須です",
		},
		{
			name: "name too long",
			config: ProjectConfig{
				Name: strings.Repeat("a", 101), // 101文字
			},
			wantErr: true,
			errMsg:  "プロジェクト名は最大100文字までです",
		},
		{
			name: "description too long",
			config: ProjectConfig{
				Name:        "test-project",
				Description: strings.Repeat("a", 501), // 501文字
			},
			wantErr: true,
			errMsg:  "説明は最大500文字までです",
		},
		{
			name: "validate edge case",
			config: ProjectConfig{
				Name:        strings.Repeat("a", 100), // ちょうど100文字
				Description: strings.Repeat("a", 500), // ちょうど500文字
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProjectConfig_GetGitHubCreateCommand(t *testing.T) {
	tests := []struct {
		name     string
		config   ProjectConfig
		expected []string
	}{
		{
			name: "with template and private",
			config: ProjectConfig{
				Name:        "test-project",
				Description: "テストプロジェクト",
				Template: &Template{
					FullName: "user/template-repo",
				},
				IsPrivate: true,
			},
			expected: []string{
				"repo", "create", "test-project",
				"--template", "user/template-repo",
				"--description", "テストプロジェクト",
				"--private",
				"--clone",
			},
		},
		{
			name: "no template, public",
			config: ProjectConfig{
				Name:      "test-project",
				IsPrivate: false,
			},
			expected: []string{
				"repo", "create", "test-project",
				"--public",
				"--clone",
			},
		},
		{
			name: "empty description",
			config: ProjectConfig{
				Name:        "test-project",
				Description: "",
				IsPrivate:   true,
			},
			expected: []string{
				"repo", "create", "test-project",
				"--private",
				"--clone",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetGitHubCreateCommand()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectConfig_GetDIsplaySummary(t *testing.T) {
	template := &Template{
		FullName: "user/awesome-template",
	}

	config := ProjectConfig{
		Name:         "test-project",
		Description:  "テストプロジェクト",
		Template:     template,
		CreateGitHub: true,
		IsPrivate:    true,
		LocalPath:    "./test-project",
	}

	summary := config.GetDisplaySummary()

	require.Len(t, summary, 5)
	assert.Contains(t, summary[0], "test-project")
	assert.Contains(t, summary[1], "テストプロジェクト")
	assert.Contains(t, summary[2], "user/awesome-template")
	assert.Contains(t, summary[3], "プライベート")
	assert.Contains(t, summary[4], "./test-project")
}

func BenchmarkProjectConfig_Validate(b *testing.B) {
	config := ProjectConfig{
		Name:        "test-project",
		Description: "テストプロジェクト",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}
