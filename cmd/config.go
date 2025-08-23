package cmd

import (
	"fmt"

	"github.com/Yuki-Sakaguchi/gh-wizard/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display and edit configuration",
	Long:  "Display or edit gh-wizard configuration settings",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Configuration load error: %v\n", err)
			return
		}

		fmt.Println("🔧 gh-wizard Configuration")
		fmt.Println("========================")
		fmt.Printf("Default Visibility: %s\n", map[bool]string{true: "Private", false: "Public"}[cfg.DefaultPrivate])
		fmt.Printf("Default Clone: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[cfg.DefaultClone])
		fmt.Printf("Auto Add README: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[cfg.DefaultAddRemote])
		fmt.Printf("Cache Timeout: %d minutes\n", cfg.CacheTimeout)
		fmt.Printf("Theme: %s\n", cfg.Theme)

		if len(cfg.RecentTemplates) > 0 {
			fmt.Println("\nRecent Templates")
			for i, template := range cfg.RecentTemplates {
				fmt.Printf("  %d. %s\n", i+1, template)
			}
		}

		configPath, _ := config.GetConfigPath()
		fmt.Printf("\nConfiguration file: %s\n", configPath)
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetDefault()
		if err := cfg.Save(); err != nil {
			fmt.Printf("Configuration file creation error: %v\n", err)
			return
		}

		configPath, _ := config.GetConfigPath()
		fmt.Printf("✅ Configuration file created: %s\n", configPath)
		fmt.Printf("To edit settings, please edit the above file directly.")
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)
}
