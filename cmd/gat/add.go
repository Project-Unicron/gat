package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/platform"
	"gat/pkg/ssh"
	"strings"

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
	authMethod  string
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

		// Determine initial auth method based on flags if provided
		initialAuthMethod := strings.ToLower(authMethod)
		// Note: Validation of initialAuthMethod happens later if creating new or explicitly set

		// Load configuration
		validConfig, validationErrors, ioErr := config.LoadConfig()
		if ioErr != nil {
			return ioErr
		}
		if len(validationErrors) > 0 {
			fmt.Println(color.YellowString("\n⚠️ Found configuration issues with other profiles:"))
			for name, err := range validationErrors {
				fmt.Printf(color.YellowString("   - Profile [%s]: %v\n"), name, err)
			}
			fmt.Println()
		}

		// --- Logic to either update existing or create new profile ---
		var profileToSave config.Profile
		var effectiveAuthMethod string
		var isUpdate bool

		existingProfile, exists := validConfig.Profiles[profileName]

		if exists && overwrite {
			isUpdate = true
			fmt.Printf("🔄 Updating existing profile: %s\n", profileName)
			profileToSave = existingProfile // Start with existing values

			// Update fields based on flags that were explicitly set
			if cmd.Flags().Changed("username") {
				// Validate username format using the regex from config package
				if !config.ValidGitHubUsernameRegex.MatchString(username) {
					return fmt.Errorf("❌ invalid username format: '%s'", username)
				}
				profileToSave.Username = username
			}
			if cmd.Flags().Changed("email") {
				// Validate email format using the regex from config package
				if !config.ValidEmailRegex.MatchString(email) {
					// Allow potentially invalid emails but warn
					fmt.Printf(color.YellowString("⚠️ Warning: Updating profile [%s] with potentially invalid email format: %s\n"), profileName, email)
				}
				profileToSave.Email = email
			}
			if cmd.Flags().Changed("platform") {
				platformID = strings.ToLower(platformID)
				// Validate platform by checking if it exists in the registry
				reg := platform.NewRegistry()
				if _, err := reg.GetPlatform(platformID); err != nil {
					return fmt.Errorf("❌ invalid platform ID '%s': %w", platformID, err)
				}
				profileToSave.Platform = platformID
			}
			if cmd.Flags().Changed("host") {
				profileToSave.Host = host
			}
			if cmd.Flags().Changed("ssh-identity") {
				profileToSave.SSHIdentity = sshIdentity
			}

			// Determine effective auth method for update
			if cmd.Flags().Changed("auth-method") {
				effectiveAuthMethod = initialAuthMethod
				if effectiveAuthMethod != "ssh" && effectiveAuthMethod != "https" {
					return fmt.Errorf("❌ invalid auth_method '%s'. Must be 'ssh' or 'https'", authMethod)
				}
			} else if cmd.Flags().Changed("ssh-identity") {
				// If ssh key changed, default to ssh unless token was also changed
				if !cmd.Flags().Changed("token") {
					effectiveAuthMethod = "ssh"
				} else {
					effectiveAuthMethod = profileToSave.AuthMethod // Keep original if both changed?
				}
			} else if cmd.Flags().Changed("token") {
				// If token changed, default to https unless ssh key was also changed
				if !cmd.Flags().Changed("ssh-identity") {
					effectiveAuthMethod = "https"
				} else {
					effectiveAuthMethod = profileToSave.AuthMethod // Keep original if both changed?
				}
			} else {
				effectiveAuthMethod = profileToSave.AuthMethod // Keep existing if nothing related changed
			}
			profileToSave.AuthMethod = effectiveAuthMethod

			// Handle token update
			if cmd.Flags().Changed("token") {
				profileToSave.SetToken(token, validConfig.StoreEncrypted, validConfig.Salt)
			}

		} else {
			isUpdate = false
			// Creating a new profile or adding without overwrite
			if exists && !overwrite {
				return fmt.Errorf("❌ profile [%s] already exists. Use --overwrite to replace it", profileName)
			}

			// Validate required flags for new profile
			if !cmd.Flags().Changed("username") {
				return fmt.Errorf("❌ --username is required when adding a new profile")
			}
			if !cmd.Flags().Changed("email") {
				return fmt.Errorf("❌ --email is required when adding a new profile")
			}
			// Validate username format for new profile
			if !config.ValidGitHubUsernameRegex.MatchString(username) {
				return fmt.Errorf("❌ invalid username format: '%s'", username)
			}
			// Validate email format for new profile
			if !config.ValidEmailRegex.MatchString(email) {
				// Allow potentially invalid emails but warn
				fmt.Printf(color.YellowString("⚠️ Warning: Adding profile [%s] with potentially invalid email format: %s\n"), profileName, email)
			}

			// Determine effective auth method for new profile
			if initialAuthMethod == "" {
				if sshIdentity != "" {
					effectiveAuthMethod = "ssh"
				} else {
					effectiveAuthMethod = "https"
				}
				fmt.Printf("ℹ️ Auth method not specified, defaulting to '%s'\n", effectiveAuthMethod)
			} else {
				effectiveAuthMethod = initialAuthMethod
				if effectiveAuthMethod != "ssh" && effectiveAuthMethod != "https" {
					return fmt.Errorf("❌ invalid auth_method '%s'. Must be 'ssh' or 'https'", authMethod)
				}
			}

			// Validate platform if provided for new profile
			platformID = strings.ToLower(platformID)
			if cmd.Flags().Changed("platform") {
				reg := platform.NewRegistry()
				if _, err := reg.GetPlatform(platformID); err != nil {
					return fmt.Errorf("❌ invalid platform ID '%s': %w", platformID, err)
				}
			} else {
				platformID = "github" // Default platform if not specified for new profile
			}

			// Create the new profile struct from flags
			profileToSave = config.Profile{
				Username:    username,
				Email:       email,
				SSHIdentity: sshIdentity,
				Platform:    platformID,
				Host:        host,
				AuthMethod:  effectiveAuthMethod,
			}
			// Set token only if provided for new profile
			if cmd.Flags().Changed("token") {
				profileToSave.SetToken(token, validConfig.StoreEncrypted, validConfig.Salt)
			}
		}

		// Add or update the profile in the config map
		// AddProfile now implicitly handles the overwrite logic based on the flag
		if err := config.AddProfile(&validConfig, profileName, profileToSave, overwrite); err != nil {
			return err
		}

		// Set as current only if adding the very first profile
		if !isUpdate && len(validConfig.Profiles) == 1 {
			validConfig.Current = profileName
			fmt.Printf("✅ Set as current profile: %s\n", profileName)
		}

		// Save the updated configuration
		if err := config.SaveConfig(&validConfig); err != nil {
			return err
		}

		// Set up SSH configuration if requested AND auth method is SSH
		// Use profileToSave here as it contains the final state
		if setupSSH && profileToSave.SSHIdentity != "" && profileToSave.AuthMethod == "ssh" {
			fmt.Println("🔐 Setting up SSH configuration...")
			if err := ssh.UpdateSSHConfig(profileToSave.Platform, profileName, profileToSave.SSHIdentity); err != nil {
				fmt.Printf(color.YellowString("⚠️ Warning: Failed to update SSH config: %v\n"), err)
			}
		}

		// Print success message (use profileToSave for final values)
		fmt.Printf("✅ Added/Updated profile: %s (%s on %s, auth: %s)\n",
			color.GreenString(profileName),
			color.CyanString(profileToSave.Username),
			color.MagentaString(profileToSave.Platform),
			color.BlueString(profileToSave.AuthMethod))

		// Show reminder to switch if the added/updated profile is not the current one
		if validConfig.Current != profileName {
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
	addCmd.Flags().StringVar(&token, "token", "", "Git personal access token (used for HTTPS)")
	addCmd.Flags().StringVar(&sshIdentity, "ssh-identity", "", "Path to SSH identity file (used for SSH)")
	addCmd.Flags().StringVar(&platformID, "platform", "github", "Git platform (e.g., github, gitlab, bitbucket)")
	addCmd.Flags().StringVar(&host, "host", "", "Custom hostname for self-hosted instances")
	addCmd.Flags().StringVar(&authMethod, "auth-method", "", "Authentication method ('ssh' or 'https'). Defaults based on --ssh-identity.")
	addCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite profile if it already exists")
	addCmd.Flags().BoolVar(&setupSSH, "setup-ssh", true, "Set up SSH host alias in ~/.ssh/gat_config if using SSH auth method")

	// Mark required flags - REMOVED these as validation is handled inside RunE
	// addCmd.MarkFlagRequired("username")
	// addCmd.MarkFlagRequired("email")
	// Note: We don't require token or ssh-identity, but auth method choice implies one is needed.
}
