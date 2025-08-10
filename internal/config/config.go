package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション設定を表す
type Config struct {
	DefaultPrivate   bool     `yaml:"default_private"`
	DefaultClone     bool     `yaml:"default_clone"`
	DefaultAddRemote bool     `yaml:"default_add_remote"`
	CacheTimeout     int      `yaml:"cache_timeout"`
	Theme            string   `yaml:"theme"`
	RecentTemplates  []string `yaml:"recent_templates"`
}

// GetConfigPath は設定ファイルのパスを取得する
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gh-wizard")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil

}

// Load は設定ファイルを読み込む
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefault(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("設定ファイルの解析に失敗しました: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("設定ファイルが無効です: %w", err)
	}

	return &config, nil
}

// Save は設定ファイルを保存する
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("設定の YAML 変換に失敗しました: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("設定の保存に失敗しました: %w", err)
	}

	return nil
}

// Validate は設定値の妥当性をチェックする
func (c *Config) Validate() error {
	if c.CacheTimeout < 0 {
		return fmt.Errorf("キャッシュタイムアウトは0以上である必要があります")
	}

	if c.Theme != "" && c.Theme != "default" && c.Theme != "dark" && c.Theme != "light" {
		return fmt.Errorf("テーマは 'default', 'dark', 'light' のいずれかである必要があります")
	}

	return nil
}

// AddRecentTemplate は最近使用したテンプレートを追加する
func (c *Config) AddRecentTemplate(templateName string) {
	// すでに存在する場合は先頭に追加し直したいので削除
	for i, name := range c.RecentTemplates {
		if name == templateName {
			c.RecentTemplates = append(c.RecentTemplates[:i], c.RecentTemplates[i+1:]...)
			break
		}
	}

	// 先頭に追加
	c.RecentTemplates = append([]string{templateName}, c.RecentTemplates...)

	// 10個まで
	if len(c.RecentTemplates) > 10 {
		c.RecentTemplates = c.RecentTemplates[:10]
	}
}
