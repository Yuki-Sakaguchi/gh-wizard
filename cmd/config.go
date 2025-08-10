package cmd

import (
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "設定を表示・編集",
	Long:  "gh-wizard の設定を表示または編集します",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "現在の設定を表示",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("設定の読み込みエラー: %v\n", err)
			return
		}

		fmt.Println("🔧 gh-wizard 設定")
		fmt.Println("========================")
		fmt.Printf("デフォルト可視性: %s\n", map[bool]string{true: "Private", false: "Public"}[cfg.DefaultPrivate])
		fmt.Printf("デフォルトクローン: %s\n", map[bool]string{true: "有効", false: "無効"}[cfg.DefaultClone])
		fmt.Printf("README自動追加: %s\n", map[bool]string{true: "有効", false: "無効"}[cfg.DefaultAddRemote])
		fmt.Printf("キャッシュタイムアウト: %d 分\n", cfg.CacheTimeout)
		fmt.Printf("テーマ: %s\n", cfg.Theme)

		if len(cfg.RecentTemplates) > 0 {
			fmt.Println("\n最近使用したテンプレート")
			for i, template := range cfg.RecentTemplates {
				fmt.Printf("  %d. %s\n", i+1, template)
			}
		}

		configPath, _ := config.GetConfigPath()
		fmt.Printf("\n設定ファイル: %s\n", configPath)
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "設定ファイルを初期化",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetDefault()
		if err := cfg.Save(); err != nil {
			fmt.Printf("設定ファイルの作成エラー: %v\n", err)
			return
		}

		configPath, _ := config.GetConfigPath()
		fmt.Printf("✅ 設定ファイルを作成しました: %s\n", configPath)
		fmt.Printf("設定を編集するには、上記ファイルを直接編集してください。")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
