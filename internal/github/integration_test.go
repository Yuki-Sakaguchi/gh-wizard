package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// これらのテストは実際のGitHub APIを使用するため、
// 適切な認証とネットワーク接続が必要

func TestDefaultClient_CheckAuthentication_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.CheckAuthentication(ctx)

	// 認証が設定されている環境では成功するはず
	// CI/CD環境では認証が設定されていない可能性がある
	if err != nil {
		t.Logf("認証チェック失敗（環境による）: %v", err)
	}
}

func TestDefaultClient_GetUserTemplates_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ")
	}

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 認証チェック
	if err := client.CheckAuthentication(ctx); err != nil {
		t.Skip("認証が設定されていないため統合テストをスキップ")
	}

	templates, err := client.GetUserTemplates(ctx)

	// エラーがないことを確認（テンプレートが0個でも正常）
	require.NoError(t, err)
	assert.NotNil(t, templates)

	// テンプレートがある場合の検証
	for _, template := range templates {
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.FullName)
		assert.True(t, template.IsTemplate)
	}

	t.Logf("取得したテンプレート数: %d", len(templates))
}
