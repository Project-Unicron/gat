package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/git"
	"gat/pkg/platform"
	"gat/pkg/ssh"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "🔄 Switch to a different Git profile",
	Long: `🔄 Switches to a different Git profile.

This command updates your global Git identity (user.name, user.email).
If run inside a Git repository, it also:
- Configures the SSH agent (starts if necessary, clears old keys, adds the profile's key if AuthMethod is 'ssh').
- Updates the 'origin' remote URL to match the profile's AuthMethod ('ssh' or 'https').
- Updates stored Git credentials for HTTPS if applicable.`,
	Args: cobra.ExactArgs(1),
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
			if validationErr, isInvalid := validationErrors[profileName]; isInvalid {
				return fmt.Errorf("❌ cannot switch to profile '%s' because it failed validation: %v", profileName, validationErr)
			}
			// Otherwise, warn about other invalid profiles
			fmt.Println(color.YellowString("\n⚠️ Found configuration issues with other profiles (will be ignored):"))
			for name, err := range validationErrors {
				if name != profileName { // Don't repeat the error for the target profile
					fmt.Printf(color.YellowString("   - Profile [%s]: %v\n"), name, err)
				}
			}
			fmt.Println() // Add a newline for separation
		}

		// Get profile from the set of valid profiles
		profile, exists := validConfig.Profiles[profileName]
		if !exists {
			// If it didn't exist in validationErrors either, it's truly not found
			if _, wasInvalid := validationErrors[profileName]; !wasInvalid {
				return fmt.Errorf("❌ profile '%s' does not exist", profileName)
			} // If it *was* invalid, the error was already returned above.
			// This path shouldn't normally be reached due to the check above, but covers edge cases.
			return fmt.Errorf("❌ profile '%s' not found (it may have failed validation)", profileName)
		}

		// Get platform information
		platformID := profile.Platform // Already normalized by LoadConfig
		reg := platform.NewRegistry()
		plat, _ := reg.GetPlatform(platformID) // Ignore error, defaults handled later if needed
		platformName := platformID
		if plat != nil {
			platformName = plat.Name
		}

		// This is the line the linter was complaining about (ensure it ends with \n")
		fmt.Printf("🔄 Switching to %s profile '%s'...\n",
			color.MagentaString(platformName),
			color.GreenString(profileName))

		if dryRun {
			fmt.Println(color.YellowString("🧪 Dry run mode enabled. No changes will be made."))
			fmt.Printf("    Would set Git User: %s\n", profile.Username)
			fmt.Printf("    Would set Git Email: %s\n", profile.Email)
			fmt.Printf("    Auth Method: %s\n", profile.AuthMethod)
			if profile.AuthMethod == "ssh" {
				fmt.Printf("    Would manage SSH Key: %s\n", profile.SSHIdentity)
			} else {
				fmt.Printf("    Would use Token for HTTPS\n")
			}
			fmt.Printf("    Would ensure remote uses: %s\n", strings.ToUpper(profile.AuthMethod))
			return nil
		}

		// --- Start applying changes ---

		// 1. Set as current profile in gat config
		validConfig.Current = profileName
		// Pass address of validConfig as SaveConfig expects a pointer
		if err := config.SaveConfig(&validConfig); err != nil {
			fmt.Printf(color.RedString("  ⚠️ Failed to save current profile setting: %v\n"), err)
			// Non-fatal, continue with other steps
		}

		// 2. Update Git global identity
		if err := git.SetIdentity(profile.Username, profile.Email); err != nil {
			// This is more critical, return error
			return fmt.Errorf(color.RedString("  ❌ Failed to set Git identity: %v"), err)
		}
		fmt.Printf("  ✅ Git identity set: %s <%s>\n",
			color.CyanString(profile.Username),
			color.CyanString(profile.Email))

		// 3. Handle Auth Method specific logic
		if profile.AuthMethod == "ssh" {
			// --- SSH Logic ---
			fmt.Println(color.YellowString("  🔐 Handling SSH Configuration..."))

			// 3a. Ensure SSH agent is running
			if err := ssh.StartAgent(); err != nil {
				fmt.Printf(color.RedString("    ⚠️ Failed to start or connect to ssh-agent: %v\n"), err)
				// Non-fatal for now, maybe user handles agent manually
			} else {
				// 3b. Clear existing identities from agent
				if err := ssh.ClearIdentities(); err != nil {
					fmt.Printf(color.RedString("    ⚠️ Failed to clear identities from ssh-agent: %v\n"), err)
					// Non-fatal
				}

				// 3c. Add the profile's identity
				if profile.SSHIdentity == "" {
					fmt.Println(color.YellowString("    ⚠️ Profile '%s' uses SSH but has no SSH identity configured."), profileName)
				} else {
					// Check if identity file exists first
					exists, checkErr := ssh.CheckSSHIdentity(profile.SSHIdentity)
					if checkErr != nil {
						fmt.Printf(color.RedString("    ⚠️ Error checking SSH identity file '%s': %v\n"), profile.SSHIdentity, checkErr)
					} else if !exists {
						fmt.Printf(color.RedString("    ⚠️ SSH identity file not found: %s\n"), profile.SSHIdentity)
						fmt.Println(color.YellowString("      💡 Please ensure the key exists or update the profile."))
					} else {
						// Add identity to agent
						if err := ssh.AddIdentity(profile.SSHIdentity); err != nil {
							fmt.Printf(color.RedString("    ❌ Failed to add SSH identity '%s' to agent: %v\n"), profile.SSHIdentity, err)
							// Consider this potentially fatal? Or just warn? Warn for now.
						} else {
							fmt.Printf("    ✅ SSH identity loaded: %s\n", color.CyanString(profile.SSHIdentity))
						}
					}
				}
			}
			// 3d. Ensure SSH config includes host alias (done by 'add' or manually)
			// We assume the host alias config is correct here, but maybe add a check later?
			// ssh.ConfigureSSH(platformID, profileName, profile.SSHIdentity) // Re-running this might be too aggressive

		} else {
			// --- HTTPS Logic ---
			fmt.Println(color.YellowString("  🔑 Handling HTTPS Configuration..."))
			// 3e. Update Git credentials (uses token)
			if profile.GetToken() == "" {
				fmt.Println(color.YellowString("    ⚠️ Profile '%s' uses HTTPS but has no token configured."), profileName)
				fmt.Println(color.YellowString("      💡 Git might prompt for credentials manually."))
			} else {
				if err := git.UpdateGitCredentials(&profile); err != nil {
					fmt.Printf(color.RedString("    ⚠️ Failed to update Git credentials: %v\n"), err)
					// Non-fatal, maybe user uses a different credential method
				} else {
					fmt.Printf("    ✅ Git credentials updated for %s\n", color.CyanString(profile.Username))
				}
			}
		}

		// 4. Update Git remote URL if in a repository
		if git.IsInGitRepo() {
			fmt.Println(color.YellowString("  🔗 Handling Git Remote URL..."))
			finalURL, err := git.RewriteRemote(&profile, profileName)
			if err != nil {
				fmt.Printf(color.RedString("    ⚠️ Failed to rewrite remote URL: %v\n"), err)
				// Non-fatal
			} else if finalURL != "" {
				fmt.Printf("    ✅ Remote 'origin' set to use %s: %s\n",
					color.CyanString(strings.ToUpper(profile.AuthMethod)),
					color.CyanString(finalURL))
			} else {
				// This case happens if RewriteRemote couldn't get the current URL
				fmt.Println(color.YellowString("    ℹ️ Skipping remote rewrite (could not determine current remote)."))
			}
		} else {
			fmt.Println(color.YellowString("  ℹ️ Not inside a Git repository, skipping remote URL update."))
		}

		// --- End applying changes ---

		fmt.Println(color.GreenString("\n✅ Switched successfully to profile: %s", profileName))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)

	switchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the switch without making changes")
}
