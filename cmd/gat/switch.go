package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/git"
	"gat/pkg/platform"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	useSSH   bool
	useHTTPS bool
	dryRun   bool
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "🔄 Switch to a different Git profile",
	Long:  `🔄 Switches to a different Git profile and updates Git config accordingly.`,
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

		// Get profile
		profile, exists := cfg.Profiles[profileName]
		if !exists {
			return fmt.Errorf("❌ profile '%s' does not exist", profileName)
		}

		// Get platform information
		platformID := profile.GetPlatform()
		reg := platform.NewRegistry()
		plat, err := reg.GetPlatform(platformID)
		if err != nil {
			// Not a fatal error, we'll use defaults
			fmt.Printf("⚠️ Unknown platform '%s', using defaults\n", platformID)
		} else {
			fmt.Printf("🔄 Switching to %s profile '%s'\n",
				color.MagentaString(plat.Name),
				color.GreenString(profileName))
		}

		if dryRun {
			fmt.Println("🧪 Dry run mode, no changes will be made")
		} else {
			// Set as current profile
			cfg.Current = profileName

			// Save configuration
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}

			// Update Git identity
			if err := git.SetIdentity(profile.Username, profile.Email); err != nil {
				return err
			}
			fmt.Printf("✅ Updated Git identity to %s <%s>\n",
				color.CyanString(profile.Username),
				color.CyanString(profile.Email))

			// Update Git credentials
			if profile.GetToken() != "" {
				if err := git.UpdateGitCredentials(&profile); err != nil {
					fmt.Printf("⚠️ Could not update Git credentials: %v\n", err)
				} else {
					fmt.Printf("✅ Updated Git credentials for %s\n",
						color.CyanString(profile.Username))
				}
			}

			// Update Git remote if in a Git repository
			if git.IsInGitRepo() {
				// Determine which protocol to use (SSH or HTTPS)
				var protocol string
				if useSSH && useHTTPS {
					return fmt.Errorf("❌ cannot use both --ssh and --https flags together")
				} else if useSSH {
					protocol = "SSH"
				} else if useHTTPS {
					protocol = "HTTPS"
				} else {
					// Default to SSH if SSH identity is set
					if profile.SSHIdentity != "" {
						protocol = "SSH"
					} else {
						protocol = "HTTPS"
					}
				}

				// Update the remote protocol
				if protocol == "SSH" {
					if err := git.UpdateRemoteProtocol(true, &profile, profileName); err != nil {
						fmt.Printf("⚠️ Could not update Git remote to SSH: %v\n", err)
					} else {
						fmt.Println("✅ Updated Git remote to use SSH")
					}
				} else {
					if err := git.UpdateRemoteProtocol(false, &profile, profileName); err != nil {
						fmt.Printf("⚠️ Could not update Git remote to HTTPS: %v\n", err)
					} else {
						fmt.Println("✅ Updated Git remote to use HTTPS")
					}
				}
			}
		}

		// Print success message
		fmt.Println(color.GreenString("✅ Switched to profile: %s", profileName))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)

	// Add flags
	switchCmd.Flags().BoolVar(&useSSH, "ssh", false, "Use SSH protocol for Git operations")
	switchCmd.Flags().BoolVar(&useHTTPS, "https", false, "Use HTTPS protocol for Git operations")
	switchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the switch without making changes")
}
