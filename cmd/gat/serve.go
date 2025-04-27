package main

import (
	"fmt"
	"gat/pkg/api/graphql"
	"gat/pkg/api/rest"
	"gat/pkg/api/server"
	"gat/pkg/config"
	"gat/pkg/git"
	"gat/pkg/platform"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	apiPort int
	apiHost string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "🌐 Start a local API server",
	Long: `🌐 Start a local API server that exposes GAT functionality via REST and GraphQL.
This allows other tools and UIs to interact with GAT programmatically.

By default, the server binds to localhost:9999 for security reasons.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get config directory
		configPath, err := config.ConfigPath()
		if err != nil {
			fmt.Printf("❌ Failed to get config directory: %v\n", err)
			os.Exit(1)
		}

		// Create server configuration
		serverConfig := server.Config{
			Port:      apiPort,
			Host:      apiHost,
			ConfigDir: configPath,
		}

		// Initialize the server
		apiServer := server.NewServer(serverConfig)

		// Set up the dependencies
		configManager := config.NewManager(configPath)
		platformReg := platform.NewRegistry()
		gitManager := git.NewManager(configManager, platformReg)

		// Set up REST handlers
		restHandler := rest.NewHandler(configManager, platformReg)
		restHandler.RegisterRoutes(apiServer.GetServeMux())

		// Set up GraphQL handlers
		resolver := graphql.NewResolver(configManager, platformReg, gitManager)
		apiServer.RegisterHandler("/graphql", graphql.Handler(resolver))
		apiServer.RegisterHandler("/playground", graphql.PlaygroundHandler())

		// Start the server
		if err := apiServer.Start(); err != nil {
			fmt.Printf("❌ Failed to start server: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(color.GreenString("✅ GAT API server started on %s:%d", apiHost, apiPort))
		fmt.Println(color.CyanString("🔎 REST API available at http://%s:%d/profiles, /platforms, /doctor", apiHost, apiPort))
		fmt.Println(color.CyanString("🔮 GraphQL API available at http://%s:%d/graphql", apiHost, apiPort))
		fmt.Println(color.CyanString("🛝 GraphQL Playground at http://%s:%d/playground", apiHost, apiPort))
		fmt.Println(color.YellowString("Press Ctrl+C to stop"))

		// Set up signal handling for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		fmt.Println(color.YellowString("\nShutting down server..."))
		if err := apiServer.Stop(); err != nil {
			fmt.Printf("❌ Error stopping server: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(color.GreenString("✅ Server stopped"))
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Add flags
	serveCmd.Flags().IntVar(&apiPort, "port", 9999, "Port to run the server on")
	serveCmd.Flags().StringVar(&apiHost, "host", "localhost", "Host to bind the server to")
}
