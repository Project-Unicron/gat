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

		// Ensure config file exists
		configFilePath, err := config.ConfigFilePath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			emptyConfig := &config.Config{
				Current:  "",
				Profiles: make(map[string]config.Profile),
			}
			if err := config.SaveConfig(emptyConfig); err != nil {
				return fmt.Errorf("❌ could not create initial config file: %w", err)
			}
			fmt.Printf("✅ Created empty configuration file at %s\n\n", configFilePath)
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Get current profile
		profile, profileName, err := config.GetCurrentProfile(cfg)
		if err != nil {
			fmt.Println("⚠️ No active profile.")
			fmt.Println("👉 Use 'gat switch <name>' to activate a profile.")
			return nil
		}

		// Print profile information
		fmt.Println("🔍 Current Profile:")
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
