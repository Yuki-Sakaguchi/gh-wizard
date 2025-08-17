package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "対話式リポジトリ作成ウィザードを開始",
	Long:  "🔮 GitHub Repository Wizard",
	RunE:  runWizard,
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}

func runWizard(cmd *cobra.Command, args []string) error {
	fmt.Println("🔮 GitHub Repository Wizard")
	fmt.Println("実装準備中")

	return nil
}
