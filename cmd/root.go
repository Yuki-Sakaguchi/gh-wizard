package cmd

import (
	"fmt"
	"os"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/config"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wizard",
	Short: "🔮 GitHub Repository Wizard",
	Long:  "魔法のように簡単で直感的なGitHubリポジトリ作成ウィザード",
	Run: func(cmd *cobra.Command, args []string) {
		// 設定を読み込み
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "設定の読み込みエラー: %v\n", err)
			os.Exit(1)
		}

		// GitHub クライアントを作成
		githubClient := github.NewClient()

		// GitHub CLI の認証チェック
		if err := githubClient.IsAuthenticated(); err != nil {
			fmt.Fprintf(os.Stderr, "GitHub認証エラー: %v\n", err)
			fmt.Fprintf(os.Stderr, "'gh auth login' を実行してください。\n")
			os.Exit(1)
		}

		// デバッグフラグの取得
		debug, _ := cmd.Flags().GetBool("debug")

		// TUI アプリケーションを実行
		if err := tui.Run(cfg, githubClient, debug); err != nil {
			fmt.Fprintf(os.Stderr, "アプリケーションエラー: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().Bool("debug", false, "デバッグモードで実行")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
