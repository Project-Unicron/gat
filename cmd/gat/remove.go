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

		// Load configuration, print warnings for invalid profiles but proceed
		validConfig, validationErrors, ioErr := config.LoadConfig()
		if ioErr != nil {
			return ioErr // Handle file I/O or parsing errors first
		}
		if len(validationErrors) > 0 {
			// Check if the target profile itself failed validation
			if _, isInvalid := validationErrors[profileName]; isInvalid {
				return fmt.Errorf("❌ cannot remove profile '%s' because it failed validation: %v", profileName, validationErrors[profileName])
			}
			// Otherwise, warn about other invalid profiles
			fmt.Println(color.YellowString("\n⚠️ Found configuration issues with other profiles (will be ignored):"))
			for name, err := range validationErrors {
				if name != profileName { // Don't repeat the error for the target profile if it was valid
					fmt.Printf(color.YellowString("   - Profile [%s]: %v\n"), name, err)
				}
			}
			fmt.Println() // Add a newline for separation
		}

		// Check if profile exists in the set of valid profiles
		if _, exists := validConfig.Profiles[profileName]; !exists {
			// If it didn't exist in validationErrors either, it's truly not found
			if _, wasInvalid := validationErrors[profileName]; !wasInvalid {
				return fmt.Errorf("❌ profile '%s' does not exist", profileName)
			} // If it *was* invalid, the error was already returned above.
			// This path shouldn't normally be reached due to the check above, but covers edge cases.
			return fmt.Errorf("❌ profile '%s' not found (it may have failed validation)", profileName)
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
		if err := config.RemoveProfile(&validConfig, profileName, noBackup); err != nil {
			return err
		}

		// Save configuration
		if err := config.SaveConfig(&validConfig); err != nil {
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
