package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/git"
	"gat/pkg/platform"
	"gat/pkg/ssh"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "🩺 Diagnose Git configuration issues",
	Long:  `🩺 Diagnose Git configuration issues and provides solutions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Main title
		fmt.Println(color.CyanString("🩺 Git Account Doctor"))
		fmt.Println(color.CyanString("==================="))

		// Current Git identity
		fmt.Println("\n" + color.YellowString("🔍 Current Git Identity:"))
		identity, err := git.DiagnoseGitIdentity()
		if err != nil {
			return err
		}

		// Display identity info
		fmt.Printf("  Username: %s\n", formatValue(identity["username"]))
		fmt.Printf("  Email: %s\n", formatValue(identity["email"]))
		fmt.Printf("  Credential Helper: %s\n", formatValue(identity["credential_helper"]))

		// Display SSH setup
		sshConfigured := identity["ssh_configured"] == "true"
		fmt.Printf("  SSH Configured: %s\n", formatBool(sshConfigured))

		if !sshConfigured {
			fmt.Printf("  %s SSH setup not found or incomplete\n", color.RedString("⚠️"))
			fmt.Printf("  %s Run 'git switch <profile> --ssh' to configure SSH\n", color.YellowString("💡"))
		}

		// Current repository info
		inRepo := identity["in_git_repo"] == "true"
		if inRepo {
			fmt.Printf("  Current Repository: %s\n", color.GreenString("✓"))
			fmt.Printf("  Remote URL: %s\n", formatValue(identity["remote_url"]))
			fmt.Printf("  Protocol: %s\n", formatValue(identity["protocol"]))
		} else {
			fmt.Printf("  Current Repository: %s (not in a Git repository)\n", color.YellowString("⚠️"))
		}

		// Load config file
		fmt.Println("\n" + color.YellowString("🔍 Configuration:"))
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Check config path
		configPath, err := config.ConfigFilePath()
		if err != nil {
			return err
		}
		fmt.Printf("  Config File: %s\n", configPath)

		// Check file permissions
		if info, err := os.Stat(configPath); err == nil {
			mode := info.Mode().Perm()
			if mode&0077 != 0 {
				fmt.Printf("  %s Config file permissions are too open: %s\n", color.RedString("⚠️"), mode)
				fmt.Printf("  %s Run 'chmod 600 %s' to secure your config\n", color.YellowString("💡"), configPath)
			} else {
				fmt.Printf("  File Permissions: %s\n", color.GreenString("✓"))
			}
		}

		// Check config security settings
		fmt.Printf("  Tokens Stored: %s\n", formatBool(!cfg.NoStoreTokens))
		fmt.Printf("  Token Encryption: %s\n", formatBool(cfg.StoreEncrypted))

		if !cfg.StoreEncrypted && !cfg.NoStoreTokens {
			fmt.Printf("  %s Tokens are stored in plaintext\n", color.RedString("⚠️"))
			fmt.Printf("  %s Consider enabling encryption or not storing tokens\n", color.YellowString("💡"))
		}

		// Profile information
		fmt.Println("\n" + color.YellowString("🔍 Profiles:"))
		if len(cfg.Profiles) == 0 {
			fmt.Printf("  %s No profiles configured\n", color.YellowString("⚠️"))
			fmt.Printf("  %s Run 'gat add <name> --username <name> --email <email>' to add a profile\n", color.YellowString("💡"))
		} else {
			fmt.Printf("  Profiles: %d\n", len(cfg.Profiles))
			fmt.Printf("  Current: %s\n", formatValue(cfg.Current))

			// Get a sorted list of profile names
			var profileNames []string
			for name := range cfg.Profiles {
				profileNames = append(profileNames, name)
			}
			sort.Strings(profileNames)

			// Platform registry for validation
			reg := platform.NewRegistry()
			platformList := reg.ListPlatforms()
			platformIDs := make(map[string]bool)
			for _, plat := range platformList {
				platformIDs[plat.ID] = true
			}

			// Check each profile
			for _, name := range profileNames {
				profile := cfg.Profiles[name]
				fmt.Printf("\n  Profile: %s\n", color.GreenString(name))
				fmt.Printf("    Username: %s\n", formatValue(profile.Username))
				fmt.Printf("    Email: %s\n", formatValue(profile.Email))

				// Platform info
				platformID := profile.Platform
				if platformID == "" {
					platformID = "github" // Default to github for backward compatibility
				}
				fmt.Printf("    Platform: %s\n", formatValue(platformID))

				// Check if platform is valid
				if !platformIDs[platformID] {
					fmt.Printf("    %s Unknown platform '%s'\n", color.RedString("⚠️"), platformID)
					fmt.Printf("    %s Add this platform to ~/.gat/platforms.yaml\n", color.YellowString("💡"))
				}

				// Check host info
				if profile.Host != "" {
					fmt.Printf("    Host: %s\n", formatValue(profile.Host))
					// Check for duplicate host
					for otherName, otherProfile := range cfg.Profiles {
						otherPlatformID := otherProfile.Platform
						if otherPlatformID == "" {
							otherPlatformID = "github" // Default for backward compatibility
						}
						if otherName != name && otherProfile.Host == profile.Host && otherPlatformID == platformID {
							fmt.Printf("    %s Duplicate host with profile '%s'\n", color.RedString("⚠️"), otherName)
						}
					}
				} else {
					// Get default host from platform
					plat, err := reg.GetPlatform(platformID)
					if err == nil {
						fmt.Printf("    Host: %s (default)\n", formatValue(plat.DefaultHost))
					}
				}

				// Token info (securely)
				hasToken := profile.GetToken() != ""
				fmt.Printf("    Token: %s\n", formatBool(hasToken))

				// SSH identity info
				hasSSH := profile.SSHIdentity != ""
				fmt.Printf("    SSH Identity: %s\n", formatSSHIdentity(profile.SSHIdentity, hasSSH))

				// Check if SSH identity exists
				if hasSSH {
					exists, err := ssh.CheckSSHIdentity(profile.SSHIdentity)
					if err != nil {
						fmt.Printf("    %s Could not check SSH identity: %v\n", color.RedString("⚠️"), err)
					} else if !exists {
						fmt.Printf("    %s SSH identity file not found: %s\n", color.RedString("⚠️"), profile.SSHIdentity)
						fmt.Printf("    %s Make sure the SSH key exists or update the profile\n", color.YellowString("💡"))
					}
				}

				// Check for protocol conflict (token but no SSH, or SSH but no token)
				if !hasToken && !hasSSH {
					fmt.Printf("    %s No authentication method configured\n", color.RedString("⚠️"))
					fmt.Printf("    %s Add a token or SSH identity to this profile\n", color.YellowString("💡"))
				}
			}
		}

		// SSH configuration
		fmt.Println("\n" + color.YellowString("🔍 SSH Configuration:"))
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Check SSH config files
		sshConfigPath := filepath.Join(homeDir, ".ssh", "config")
		gatConfigPath := filepath.Join(homeDir, ".ssh", "gat_config")

		// Main SSH config
		_, err = os.Stat(sshConfigPath)
		if os.IsNotExist(err) {
			fmt.Printf("  %s SSH config file not found\n", color.RedString("⚠️"))
			fmt.Printf("  %s Run 'gat switch <profile> --ssh' to create it\n", color.YellowString("💡"))
		} else if err != nil {
			fmt.Printf("  %s Could not check SSH config: %v\n", color.RedString("⚠️"), err)
		} else {
			fmt.Printf("  SSH Config: %s\n", sshConfigPath)

			// Check for include line
			data, err := os.ReadFile(sshConfigPath)
			if err != nil {
				fmt.Printf("  %s Could not read SSH config: %v\n", color.RedString("⚠️"), err)
			} else {
				if strings.Contains(string(data), "Include ~/.ssh/gat_config") {
					fmt.Printf("  Include Line: %s\n", color.GreenString("✓"))
				} else {
					fmt.Printf("  %s SSH config does not include gat_config\n", color.RedString("⚠️"))
					fmt.Printf("  %s Add 'Include ~/.ssh/gat_config' to your SSH config\n", color.YellowString("💡"))
				}
			}
		}

		// gat_config
		_, err = os.Stat(gatConfigPath)
		if os.IsNotExist(err) {
			fmt.Printf("  %s gat SSH config file not found\n", color.RedString("⚠️"))
			fmt.Printf("  %s Run 'gat switch <profile> --ssh' to create it\n", color.YellowString("💡"))
		} else if err != nil {
			fmt.Printf("  %s Could not check gat SSH config: %v\n", color.RedString("⚠️"), err)
		} else {
			fmt.Printf("  gat SSH Config: %s\n", gatConfigPath)

			// Check file permissions
			if info, err := os.Stat(gatConfigPath); err == nil {
				mode := info.Mode().Perm()
				if mode&0077 != 0 {
					fmt.Printf("  %s gat SSH config permissions are too open: %s\n", color.RedString("⚠️"), mode)
					fmt.Printf("  %s Run 'chmod 600 %s' to secure your config\n", color.YellowString("💡"), gatConfigPath)
				} else {
					fmt.Printf("  File Permissions: %s\n", color.GreenString("✓"))
				}
			}
		}

		// Final summary
		fmt.Println("\n" + color.YellowString("🔍 Summary:"))
		if len(cfg.Profiles) == 0 {
			fmt.Printf("  %s No profiles configured. Add at least one profile to get started.\n", color.RedString("⚠️"))
		} else if cfg.Current == "" {
			fmt.Printf("  %s No active profile. Switch to a profile using 'gat switch <name>'.\n", color.YellowString("⚠️"))
		} else {
			fmt.Printf("  %s Using profile '%s' with %s on %s\n", color.GreenString("✓"), cfg.Current,
				formatValue(cfg.Profiles[cfg.Current].Username),
				formatValue(getPlatformID(cfg.Profiles[cfg.Current])))
		}

		return nil
	},
}

// getPlatformID is a helper to get the platform ID from a profile
func getPlatformID(profile config.Profile) string {
	if profile.Platform == "" {
		return "github" // Default for backward compatibility
	}
	return profile.Platform
}

// formatValue formats a value for display, handling empty strings
func formatValue(value string) string {
	if value == "" {
		return color.RedString("<not set>")
	}
	return value
}

// formatBool formats a boolean value for display
func formatBool(value bool) string {
	if value {
		return color.GreenString("✓")
	}
	return color.RedString("✗")
}

// formatSSHIdentity formats an SSH identity path for display
func formatSSHIdentity(path string, hasSSH bool) string {
	if !hasSSH {
		return color.RedString("<not set>")
	}
	return path
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
