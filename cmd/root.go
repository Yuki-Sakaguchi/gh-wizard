package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wizard",
	Short: "🔮 GitHub Repository Wizard",
	Long:  "魔法のように簡単で直感的なGitHubリポジトリ作成ウィザード",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔮 gh-wizard - GitHub Repository Wizard")
		fmt.Println("まだ実装中です...")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
