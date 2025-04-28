package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/git"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "🔍 Show current GitHub profile status",
	Long:  `🔍 Displays information about the current active GitHub profile and repository settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config directory exists
		configPath, err := config.ConfigPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := os.MkdirAll(configPath, 0755); err != nil {
				return fmt.Errorf("❌ could not create config directory: %w", err)
			}
			fmt.Printf("✅ Created configuration directory at %s\n\n", configPath)
		}

		// Load configuration, print warnings for invalid profiles but proceed
		validConfig, validationErrors, ioErr := config.LoadConfig()
		if ioErr != nil {
			return ioErr // Handle file I/O or parsing errors first
		}
		if len(validationErrors) > 0 {
			fmt.Println(color.YellowString("\n⚠️ Found configuration issues with some profiles:"))
			for name, err := range validationErrors {
				fmt.Printf(color.YellowString("   - Profile [%s]: %v\n"), name, err)
			}
			fmt.Println() // Add a newline for separation
		}

		// Get current profile based on the valid configuration
		// Pass address of validConfig as GetCurrentProfile expects a pointer
		profile, profileName, err := config.GetCurrentProfile(&validConfig)
		if err != nil {
			// This handles both "Current" being empty and "Current" pointing to an invalid profile
			fmt.Println("⚠️ No active profile set or the active profile is invalid.")
			fmt.Println("👉 Use 'gat switch <name>' to activate a valid profile.")
			return nil
		}

		// Print profile information
		fmt.Println("�� Current Profile:")
		fmt.Printf("   Name: %s\n", color.GreenString(profileName))
		fmt.Printf("   👤 Username: %s\n", profile.Username)
		fmt.Printf("   📧 Email: %s\n", profile.Email)

		if profile.SSHIdentity != "" {
			fmt.Printf("   🔑 SSH Identity: %s\n", profile.SSHIdentity)
		}

		// Check Git repository information
		if git.IsInGitRepo() {
			fmt.Println()
			fmt.Println("📁 Git Repository:")

			// Get and display remote URL
			remoteURL, err := git.GetCurrentRemoteURL()
			if err != nil {
				fmt.Println("   ⚠️ No remote URL found.")
			} else {
				fmt.Printf("   🔗 Remote URL: %s\n", remoteURL)

				// Display protocol
				if git.IsSSHRemote(remoteURL) {
					fmt.Printf("   🚀 Protocol: %s\n", color.CyanString("SSH"))
				} else {
					fmt.Printf("   🌐 Protocol: %s\n", color.CyanString("HTTPS"))
				}
			}
		} else {
			fmt.Println()
			fmt.Println("⚠️ Not in a Git repository.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
