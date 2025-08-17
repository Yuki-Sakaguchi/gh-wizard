package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "有効な設定",
			config: Config{
				CacheTimeout: 30,
				Theme:        "default",
			},
			wantErr: false,
		},
		{
			name: "無効なキャッシュタイムアウト",
			config: Config{
				CacheTimeout: -1,
				Theme:        "default",
			},
			wantErr: true,
		},
		{
			name: "無効なテーマ",
			config: Config{
				CacheTimeout: 30,
				Theme:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// 一時ディレクトリでテスト
	tempDir, err := os.MkdirTemp("", "gh-wizard-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 元の HOME 環境変数を保存
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// テスト用の HOME ディレクトリを設定
	os.Setenv("HOME", tempDir)

	// 設定を保存
	config := GetDefault()
	config.Theme = "dark"
	config.CacheTimeout = 60

	if err := config.Save(); err != nil {
		t.Fatalf("設定の保存に失敗: %v", err)
	}

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 値の検証
	if loadedConfig.Theme != "dark" {
		t.Errorf("Theme: 期待値 'dark', 実際の値 '%s'", loadedConfig.Theme)
	}

	if loadedConfig.CacheTimeout != 60 {
		t.Errorf("CacheTimeout: 期待値 60, 実際の値 %d", loadedConfig.CacheTimeout)
	}
}

func TestConfig_AddRecentTemplate(t *testing.T) {
	config := GetDefault()

	// 新しいテンプレートを追加
	config.AddRecentTemplate("user/template1")
	assert.Len(t, config.RecentTemplates, 1)
	assert.Equal(t, "user/template1", config.RecentTemplates[0])

	// 別のテンプレートを追加
	config.AddRecentTemplate("user/template2")
	assert.Len(t, config.RecentTemplates, 2)
	assert.Equal(t, "user/template2", config.RecentTemplates[0]) // 最新が最初

	// 重複するテンプレートを追加
	config.AddRecentTemplate("user/template1")
	assert.Len(t, config.RecentTemplates, 2)                     // 長さは変わらない
	assert.Equal(t, "user/template1", config.RecentTemplates[0]) // 最前面に移動
}

func TestConfig_RecentTemplateLimit(t *testing.T) {
	config := GetDefault()

	// 11個のテンプレートを追加（制限は10個）
	for i := 0; i < 11; i++ {
		config.AddRecentTemplate(fmt.Sprintf("user/template%d", i))
	}

	assert.Len(t, config.RecentTemplates, 10)                     // 最大10個
	assert.Equal(t, "user/template10", config.RecentTemplates[0]) // 最新
	assert.Equal(t, "user/template1", config.RecentTemplates[9])  // 最古（template0は削除済み）
}
