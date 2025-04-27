package main

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/platform"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "📋 List all stored profiles",
	Long:  `📋 Lists all stored Git profiles across all platforms, highlighting the current active one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config directory and file exist
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

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		// Check if we have any profiles
		if len(cfg.Profiles) == 0 {
			fmt.Println("😶 No profiles found. Add one with 'gat add <n>'")
			return nil
		}

		// Initialize platform registry
		reg := platform.NewRegistry()

		// Get a sorted list of profile names
		var profileNames []string
		for name := range cfg.Profiles {
			profileNames = append(profileNames, name)
		}
		sort.Strings(profileNames)

		// Display profiles
		fmt.Println("📋 Git Profiles:")
		fmt.Println("--------------")

		for _, name := range profileNames {
			profile := cfg.Profiles[name]

			// Get platform name
			platformID := getPlatformID(profile)
			plat, err := reg.GetPlatform(platformID)

			var platformName string
			if err != nil {
				platformName = platformID // Use ID if platform not found
			} else {
				platformName = plat.Name
			}

			// Get host name (custom or default)
			var hostName string
			if profile.Host != "" {
				hostName = profile.Host
			} else if plat != nil {
				hostName = plat.DefaultHost
			} else {
				hostName = "unknown host"
			}

			if name == cfg.Current {
				// Current profile
				fmt.Printf("%s %s\n", color.GreenString("✅"), color.GreenString(name))
				fmt.Printf("   🌐 Platform: %s (%s)\n", platformName, hostName)
				fmt.Printf("   👤 Username: %s\n", profile.Username)
				fmt.Printf("   📧 Email: %s\n", profile.Email)
				if profile.SSHIdentity != "" {
					fmt.Printf("   🔑 SSH Key: %s\n", profile.SSHIdentity)
				}
			} else {
				// Other profiles
				fmt.Printf("⬜ %s\n", name)
				fmt.Printf("   🌐 Platform: %s (%s)\n", platformName, hostName)
				fmt.Printf("   👤 Username: %s\n", profile.Username)
				fmt.Printf("   📧 Email: %s\n", profile.Email)
				if profile.SSHIdentity != "" {
					fmt.Printf("   🔑 SSH Key: %s\n", profile.SSHIdentity)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
