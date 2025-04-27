package main

import (
	"fmt"
	"gat/pkg/platform"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Platform register flags
	platID          string
	platName        string
	platHost        string
	platSSHPrefix   string
	platHTTPSPrefix string
	platSSHUser     string
	platTokenScope  string
	platYAMLPath    string
	platForce       bool
)

// platformRegisterCmd represents the register command
var platformRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a custom Git hosting platform",
	Long: `Define a new custom Git platform for use with gat profiles.
You can register a platform using command-line flags or a YAML file.

Example YAML file format:
  name: "Gitea"
  defaultHost: "git.example.com"
  sshPrefix: "git@git.example.com:"
  httpsPrefix: "https://git.example.com/"
  sshUser: "git"
  tokenAuthScope: "git.example.com"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine if we're using YAML file or flags
		var newPlatform *platform.Platform

		if platYAMLPath != "" {
			// Load from YAML file
			data, err := os.ReadFile(platYAMLPath)
			if err != nil {
				return fmt.Errorf("❌ could not read YAML file: %w", err)
			}

			// Parse YAML
			tempPlatform := &platform.Platform{}
			if err := yaml.Unmarshal(data, tempPlatform); err != nil {
				return fmt.Errorf("❌ could not parse YAML file: %w", err)
			}

			// Set the ID from the command-line if provided
			if platID != "" {
				tempPlatform.ID = platID
			}

			// Validate required fields
			if tempPlatform.ID == "" {
				return fmt.Errorf("❌ platform ID is required (either in YAML or with --id flag)")
			}
			if tempPlatform.Name == "" || tempPlatform.DefaultHost == "" ||
				tempPlatform.SSHPrefix == "" || tempPlatform.HTTPSPrefix == "" {
				return fmt.Errorf("❌ missing required fields in YAML file (name, defaultHost, sshPrefix, httpsPrefix)")
			}

			// Set defaults for optional fields if not provided
			if tempPlatform.SSHUser == "" {
				tempPlatform.SSHUser = "git"
			}
			if tempPlatform.TokenAuthScope == "" {
				tempPlatform.TokenAuthScope = tempPlatform.DefaultHost
			}

			newPlatform = tempPlatform
		} else {
			// Validate required flags
			if platID == "" || platName == "" || platHost == "" ||
				platSSHPrefix == "" || platHTTPSPrefix == "" {
				return fmt.Errorf("❌ missing required flags (--id, --name, --host, --ssh-prefix, --https-prefix)")
			}

			// Set defaults for optional fields
			if platSSHUser == "" {
				platSSHUser = "git"
			}
			if platTokenScope == "" {
				platTokenScope = platHost
			}

			// Create new platform
			newPlatform = &platform.Platform{
				ID:             platID,
				Name:           platName,
				DefaultHost:    platHost,
				SSHPrefix:      platSSHPrefix,
				HTTPSPrefix:    platHTTPSPrefix,
				SSHUser:        platSSHUser,
				TokenAuthScope: platTokenScope,
				Custom:         true,
			}
		}

		// Get user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("❌ could not find home directory: %w", err)
		}

		// Path to custom platforms file
		configDir := filepath.Join(homeDir, ".gat")
		platformsPath := filepath.Join(configDir, "platforms.yaml")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("❌ could not create config directory: %w", err)
		}

		// Load existing platforms or create new map
		existingPlatforms := make(map[string]*platform.Platform)
		if _, err := os.Stat(platformsPath); err == nil {
			// File exists, read it
			data, err := os.ReadFile(platformsPath)
			if err != nil {
				return fmt.Errorf("❌ could not read platforms file: %w", err)
			}

			// Parse YAML
			if err := yaml.Unmarshal(data, &existingPlatforms); err != nil {
				return fmt.Errorf("❌ could not parse platforms file: %w", err)
			}

			// Check if platform already exists
			if _, exists := existingPlatforms[newPlatform.ID]; exists && !platForce {
				// Check if we're in a non-interactive environment (such as CI or tests)
				// by checking if stdin is connected to a terminal
				fileInfo, _ := os.Stdin.Stat()
				isTerminal := (fileInfo.Mode() & os.ModeCharDevice) != 0

				if isTerminal {
					// Prompt for confirmation only in interactive mode
					fmt.Printf("⚠️ Platform '%s' already exists. Overwrite? (y/N): ", newPlatform.ID)
					var input string
					fmt.Scanln(&input)
					if !strings.EqualFold(input, "y") && !strings.EqualFold(input, "yes") {
						fmt.Println("Operation cancelled.")
						return nil
					}
				} else {
					// In non-interactive mode, just return an error
					return fmt.Errorf("❌ platform '%s' already exists (use --force to overwrite)", newPlatform.ID)
				}
			}
		}

		// Add the new platform
		existingPlatforms[newPlatform.ID] = newPlatform

		// Write the platforms file
		data, err := yaml.Marshal(existingPlatforms)
		if err != nil {
			return fmt.Errorf("❌ could not marshal platforms data: %w", err)
		}

		if err := os.WriteFile(platformsPath, data, 0644); err != nil {
			return fmt.Errorf("❌ could not write platforms file: %w", err)
		}

		fmt.Printf("✅ Successfully registered platform %s (%s)\n",
			color.GreenString(newPlatform.ID),
			color.YellowString(newPlatform.DefaultHost))

		return nil
	},
}

func init() {
	platformsCmd.AddCommand(platformRegisterCmd)

	// Add flags
	platformRegisterCmd.Flags().StringVar(&platID, "id", "", "Unique platform identifier (e.g., gitea)")
	platformRegisterCmd.Flags().StringVar(&platName, "name", "", "Display name (e.g., Gitea)")
	platformRegisterCmd.Flags().StringVar(&platHost, "host", "", "Default hostname (e.g., git.example.com)")
	platformRegisterCmd.Flags().StringVar(&platSSHPrefix, "ssh-prefix", "", "SSH URL prefix (e.g., git@git.example.com:)")
	platformRegisterCmd.Flags().StringVar(&platHTTPSPrefix, "https-prefix", "", "HTTPS URL prefix (e.g., https://git.example.com/)")
	platformRegisterCmd.Flags().StringVar(&platSSHUser, "ssh-user", "git", "SSH username (defaults to 'git')")
	platformRegisterCmd.Flags().StringVar(&platTokenScope, "token-scope", "", "Token authentication scope (defaults to host)")
	platformRegisterCmd.Flags().StringVar(&platYAMLPath, "yaml", "", "Path to YAML file containing platform definition")
	platformRegisterCmd.Flags().BoolVar(&platForce, "force", false, "Overwrite existing platform without confirmation")

	// Add example usage
	platformRegisterCmd.Example = `  # Using flags
  gat platforms register --id gitea --name "Gitea" --host "git.example.com" \
    --ssh-prefix "git@git.example.com:" --https-prefix "https://git.example.com/"

  # Using a YAML file
  gat platforms register --yaml ~/my-platform.yaml --id gitea`
}
