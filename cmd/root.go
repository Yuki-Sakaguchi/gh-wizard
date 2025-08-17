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
}

func init() {
	// ここで設定を追加する
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wizard/config.yaml)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
