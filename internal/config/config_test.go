package config

import (
	"os"
	"testing"
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

func TestConfig_AddRecentTemplate(t *testing.T) {
	config := GetDefault()

	// テンプレートを追加
	config.AddRecentTemplate("template1")
	config.AddRecentTemplate("template2")
	config.AddRecentTemplate("template1")

	// 期待値: 重複は削除され、最新が先頭になるため ["template1", "template2"]
	expected := []string{"template1", "template2"}

	if len(config.RecentTemplates) != len(expected) {
		t.Errorf("期待される長さ %d, 実際の長さ %d", len(expected), len(config.RecentTemplates))
	}

	for i, want := range expected {
		if config.RecentTemplates[i] != want {
			t.Errorf("インデックス %d: 期待値 %s, 実際の値 %s", i, want, config.RecentTemplates[i])
		}
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
