package main

import (
	"fmt"
	"gat/pkg/config"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gat",
	Short: "😸 GitHub Account Tool - Manage Git identities across multiple platforms",
	Long: `
😸 GitHub Account Tool (gat)
===========================
A smart CLI for managing Git identities across GitHub, GitLab, Bitbucket, Hugging Face, and more.
Switch between profiles, update Git configs, and manage tokens with ease.

Each profile can have its own username, email, token, SSH identity, and platform.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip this initialization for help commands
		if cmd.Name() == "help" || cmd.Name() == "__help" || cmd.Name() == "__complete" {
			return nil
		}

		// Ensure config directory exists
		configPath, err := config.ConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := os.MkdirAll(configPath, 0755); err != nil {
				return fmt.Errorf("❌ could not create config directory: %w", err)
			}

			// Create empty config file
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
				fmt.Printf("✅ Initialized configuration in %s\n\n", configPath)
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		cmd.Help()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig sets up any configuration needed before running commands
func initConfig() {
	// Nothing needed here yet
}
