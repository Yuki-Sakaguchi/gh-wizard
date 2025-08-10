package cmd

import (
	"context"
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/github"
	"github.com/Yuki-Sakaguchi/gh-wizard/internal/utils"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "GitHub認証状態の確認",
	Long:  "GitHub CLI の認証状態とユーザー情報を確認します",
	Run: func(cmd *cobra.Command, args []string) {
		// GitHub CLI インストールチェック
		if err := utils.CheckGHInstalled(); err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		version, _ := utils.GitGHVersion()
		fmt.Printf("✅ GitHub CLI: %s\n", version)

		// 認証状態チェック
		client := github.NewClient()
		if err := client.IsAuthenticated(); err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		fmt.Printf("✅ GitHub認証: OK\n")

		// ユーザー情報取得
		user, err := client.GetCurrentUser()
		if err != nil {
			fmt.Printf("⚠️ ユーザー情報取得エラー: %v\n", err)
			return
		}

		fmt.Printf("👤 ユーザー: %s (%s)\n", user.Name, user.Login)

		ctx := context.Background()
		templates, err := client.GetTemplateRepositories(ctx)
		if err != nil {
			fmt.Printf("⚠️ テンプレート取得エラー: %v\n", err)
			return
		}

		fmt.Printf("📚 テンプレートリポジトリ: %d件\n", len(templates))
		if len(templates) > 0 {
			fmt.Println("  最新3件:")
			for i, template := range templates {
				if i >= 3 {
					break
				}
				fmt.Printf("    - %s (%s) ⭐%d\n", template.Name, template.Language, template.Stars)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
