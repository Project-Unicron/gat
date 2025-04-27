package main

import (
	"fmt"
	"gat/pkg/config"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	forceRemove bool
	noBackup    bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "🗑️ Remove a GitHub profile",
	Long:  `🗑️ Removes a GitHub profile from your configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		// Validate profile name for security
		if err := config.ValidateProfileName(profileName); err != nil {
			return fmt.Errorf("❌ %v", err)
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Check if profile exists
		if _, exists := cfg.Profiles[profileName]; !exists {
			return fmt.Errorf("❌ profile '%s' does not exist", profileName)
		}

		// Confirm deletion unless force flag is set
		if !forceRemove {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("Type '%s' to confirm deletion", profileName),
				AllowEdit: true,
			}

			confirmName, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("❌ prompt failed: %w", err)
			}

			if strings.TrimSpace(confirmName) != profileName {
				return fmt.Errorf("❌ profile name mismatch, deletion canceled")
			}
		}

		if !noBackup {
			fmt.Println("💾 Creating backup of profile before deletion...")
		}

		// Remove profile
		if err := config.RemoveProfile(cfg, profileName, noBackup); err != nil {
			return err
		}

		// Save configuration
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		if !noBackup {
			configDir, _ := config.ConfigPath()
			backupPath := fmt.Sprintf("%s/backups/%s.backup.json", configDir, profileName)
			fmt.Printf("💾 Profile backup created at: %s\n", backupPath)
		}

		fmt.Println(color.RedString("🗑️ Profile '%s' has been destroyed. Poof. 💨", profileName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Add flags
	removeCmd.Flags().BoolVar(&forceRemove, "force", false, "Skip confirmation prompt (useful for scripts)")
	removeCmd.Flags().BoolVar(&noBackup, "no-backup", false, "Don't create a backup of the profile before deletion")
}
