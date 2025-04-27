package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/platform"
	"gat/pkg/ssh"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	username    string
	email       string
	token       string
	sshIdentity string
	platformID  string
	host        string
	overwrite   bool
	setupSSH    bool
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "➕ Add a new Git profile",
	Long:  `➕ Adds a new Git profile with the specified credentials.`,
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

		// Check if platform is valid
		if platformID != "" {
			reg := platform.NewRegistry()
			_, err := reg.GetPlatform(platformID)
			if err != nil {
				return fmt.Errorf("❌ %v", err)
			}
		}

		// Create new profile
		profile := config.Profile{
			Username:    username,
			Email:       email,
			SSHIdentity: sshIdentity,
			Platform:    platformID,
			Host:        host,
		}

		// Set token (with encryption if enabled)
		if token != "" {
			profile.SetToken(token, cfg.StoreEncrypted, cfg.Salt)
		}

		// Add profile to config
		if err := config.AddProfile(cfg, profileName, profile, overwrite); err != nil {
			return err
		}

		// Set as current if it's the first profile
		if len(cfg.Profiles) == 1 {
			cfg.Current = profileName
			fmt.Printf("✅ Set as current profile: %s\n", profileName)
		}

		// Save configuration
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		// Set up SSH configuration if requested
		if setupSSH && sshIdentity != "" {
			fmt.Println("🔐 Setting up SSH configuration...")
			if err := ssh.UpdateSSHConfig(profile.GetPlatform(), profileName, sshIdentity); err != nil {
				return err
			}
		}

		// Print success message
		fmt.Printf("✅ Added profile: %s (%s on %s)\n",
			color.GreenString(profileName),
			color.CyanString(username),
			color.MagentaString(profile.GetPlatform()))

		// Show reminder to switch to the profile to use it
		if cfg.Current != profileName {
			fmt.Printf("\nℹ️ To use this profile, run: %s\n", color.YellowString("gat switch "+profileName))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Add flags
	addCmd.Flags().StringVar(&username, "username", "", "Git username (must begin and end with alphanumeric characters, can contain hyphens in between)")
	addCmd.Flags().StringVar(&email, "email", "", "Git email")
	addCmd.Flags().StringVar(&token, "token", "", "Git personal access token")
	addCmd.Flags().StringVar(&sshIdentity, "ssh-identity", "", "Path to SSH identity file")
	addCmd.Flags().StringVar(&platformID, "platform", "github", "Git platform (github, gitlab, bitbucket, etc.)")
	addCmd.Flags().StringVar(&host, "host", "", "Custom hostname for self-hosted instances")
	addCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite profile if it already exists")
	addCmd.Flags().BoolVar(&setupSSH, "setup-ssh", true, "Set up SSH configuration for this profile")

	// Mark required flags
	addCmd.MarkFlagRequired("username")
	addCmd.MarkFlagRequired("email")
}
