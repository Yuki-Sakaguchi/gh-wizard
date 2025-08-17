package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckGitInstalled(t *testing.T) {
	err := CheckGitInstalled()

	assert.NoError(t, err, "Gitがインストールされていることを確認する")
}

func TestGetGitVersion(t *testing.T) {
	vsersion, err := GetGitVersion()

	require.NoError(t, err, "Gitのバージョンを取得する")
	assert.Contains(t, vsersion, "git version", "Gitのバージョンが取得できることを確認する")
	assert.NotEmpty(t, vsersion, "Gitのバージョンが空ではないことを確認する")
}

func TestCheckGitVertion(t *testing.T) {
	err := CheckGitVersion()

	assert.NoError(t, err, "Gitのバージョンが要件を満たすことを確認する")
}

func TestGitCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectError bool
	}{
		{
			name:        "valid git command",
			command:     "git",
			expectError: false,
		},
		{
			name:        "invalid git command",
			command:     "nonexistent-command",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("skipping test")
		})
	}
}

func TestCheckGHInstalled(t *testing.T) {
	err := CheckGHInstalled()

	assert.NoError(t, err, "GitHub CLIがインストールされていることを確認する")
}

func TestCheckGHVersion(t *testing.T) {
	err := CheckGHVersion()

	assert.NoError(t, err, "GitHub CLIのバージョンが要件を満たすことを確認する")
}

func TestGitGHVersion(t *testing.T) {
	version, err := GitGHVersion()

	assert.NoError(t, err, "GitHub CLIのバージョンが取得できることを確認する")
	assert.NotEmpty(t, version, "GitHub CLIのバージョンが空ではないことを確認する")
}
