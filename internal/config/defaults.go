package config

// GetDefault はデフォルト設定を返す
func GetDefault() *Config {
	return &Config{
		DefaultPrivate:   true,
		DefaultClone:     true,
		DefaultAddRemote: true,
		CacheTimeout:     30,
		Theme:            "default",
		RecentTemplates:  make([]string, 0),
	}
}

// GetConfigTemplate は設定ファイルのテンプレートYAMLを返す
func GetConfigTemplate() string {
	return `# gh-wizard 設定ファイル
# 詳細: https://github.com/Yuki-Sakaguchi/gh-wizard

# デフォルト設定
default_private: true        # リポジトリをプライベートにするか
default_clone: true          # 作成後にローカルにクローンするか
default_add_readme: true     # READMEファイルを追加するか

# キャッシュ設定
cache_timeout_minutes: 30    # テンプレート一覧のキャッシュ時間（分）

# UI設定
theme: "default"             # テーマ: default, dark, light

# 最近使用したテンプレート（自動更新）
recent_templates: []
`
}
