package main

import (
	"fmt"
	"gat/pkg/platform"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// platformsCmd represents the platforms command
var platformsCmd = &cobra.Command{
	Use:   "platforms",
	Short: "🌐 Manage supported Git hosting platforms",
	Long:  `🌐 List, register, and manage Git hosting platforms supported by gat.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior is to list platforms
		listPlatformsCmd.Run(cmd, args)
	},
}

// listPlatformsCmd represents the list subcommand of platforms
var listPlatformsCmd = &cobra.Command{
	Use:   "list",
	Short: "List built-in and custom Git hosting platforms",
	Long:  `Display all supported Git hosting platforms, including built-in and custom user-defined platforms.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new platform registry
		reg := platform.NewRegistry()

		// Get all platforms
		platforms := reg.ListPlatforms()

		// Print header
		fmt.Println("🌐 Supported Git hosting platforms:")
		fmt.Println()

		// Print built-in platforms
		fmt.Println(color.CyanString("Built-in platforms:"))
		hasBuiltIn := false
		for _, plat := range platforms {
			if !plat.Custom {
				hasBuiltIn = true
				fmt.Printf("  • %s (%s) - %s\n",
					color.GreenString(plat.ID),
					color.YellowString(plat.DefaultHost),
					plat.Name)
			}
		}
		if !hasBuiltIn {
			fmt.Println("  No built-in platforms found")
		}

		// Print custom platforms
		fmt.Println()
		fmt.Println(color.CyanString("Custom platforms:"))
		hasCustom := false
		for _, plat := range platforms {
			if plat.Custom {
				hasCustom = true
				fmt.Printf("  • %s (%s) - %s\n",
					color.GreenString(plat.ID),
					color.YellowString(plat.DefaultHost),
					plat.Name)
			}
		}
		if !hasCustom {
			fmt.Println("  No custom platforms defined")
			fmt.Println()
			fmt.Println("To register a custom platform, use the command:")
			fmt.Printf("  %s\n", color.YellowString("gat platforms register --help"))
		}
	},
}

func init() {
	rootCmd.AddCommand(platformsCmd)
	platformsCmd.AddCommand(listPlatformsCmd)
}
