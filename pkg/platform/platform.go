package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Platform represents a Git hosting platform's configuration
type Platform struct {
	ID             string `yaml:"id"`             // Unique identifier (e.g., "github", "gitlab")
	Name           string `yaml:"name"`           // Display name (e.g., "GitHub", "GitLab")
	DefaultHost    string `yaml:"defaultHost"`    // Default hostname (e.g., "github.com")
	SSHPrefix      string `yaml:"sshPrefix"`      // SSH prefix (e.g., "git@github.com:")
	HTTPSPrefix    string `yaml:"httpsPrefix"`    // HTTPS prefix (e.g., "https://github.com/")
	SSHUser        string `yaml:"sshUser"`        // SSH username (typically "git")
	TokenAuthScope string `yaml:"tokenAuthScope"` // Token authentication scope (e.g., "github.com")
	Custom         bool   `yaml:"custom"`         // Whether this is a custom user-defined platform
}

// Registry holds all registered Git hosting platforms
type Registry struct {
	Platforms map[string]*Platform // Map of platform ID to Platform
}

// NewRegistry creates a new platform registry with default platforms
func NewRegistry() *Registry {
	reg := &Registry{
		Platforms: make(map[string]*Platform),
	}

	// Register default platforms
	reg.registerDefaults()

	// Load custom platforms
	if err := reg.loadCustomPlatforms(); err != nil {
		// Just log the error, don't fail
		fmt.Printf("⚠️ Warning: could not load custom platforms: %s\n", err)
	}

	return reg
}

// registerDefaults registers the default Git hosting platforms
func (r *Registry) registerDefaults() {
	defaults := []*Platform{
		{
			ID:             "github",
			Name:           "GitHub",
			DefaultHost:    "github.com",
			SSHPrefix:      "git@github.com:",
			HTTPSPrefix:    "https://github.com/",
			SSHUser:        "git",
			TokenAuthScope: "github.com",
		},
		{
			ID:             "gitlab",
			Name:           "GitLab",
			DefaultHost:    "gitlab.com",
			SSHPrefix:      "git@gitlab.com:",
			HTTPSPrefix:    "https://gitlab.com/",
			SSHUser:        "git",
			TokenAuthScope: "gitlab.com",
		},
		{
			ID:             "bitbucket",
			Name:           "Bitbucket",
			DefaultHost:    "bitbucket.org",
			SSHPrefix:      "git@bitbucket.org:",
			HTTPSPrefix:    "https://bitbucket.org/",
			SSHUser:        "git",
			TokenAuthScope: "bitbucket.org",
		},
		{
			ID:             "huggingface",
			Name:           "Hugging Face",
			DefaultHost:    "huggingface.co",
			SSHPrefix:      "git@hf.co:",
			HTTPSPrefix:    "https://huggingface.co/",
			SSHUser:        "git",
			TokenAuthScope: "huggingface.co",
		},
		{
			ID:             "azuredevops",
			Name:           "Azure DevOps",
			DefaultHost:    "dev.azure.com",
			SSHPrefix:      "git@ssh.dev.azure.com:v3/",
			HTTPSPrefix:    "https://dev.azure.com/",
			SSHUser:        "git",
			TokenAuthScope: "dev.azure.com",
		},
	}

	for _, platform := range defaults {
		r.Platforms[platform.ID] = platform
	}
}

// loadCustomPlatforms loads user-defined platforms from ~/.gat/platforms.yaml
func (r *Registry) loadCustomPlatforms() error {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %w", err)
	}

	// Path to custom platforms file
	platformsPath := filepath.Join(homeDir, ".gat", "platforms.yaml")

	// Check if the file exists
	if _, err := os.Stat(platformsPath); os.IsNotExist(err) {
		// No custom platforms file, which is fine
		return nil
	}

	// Read the file
	data, err := os.ReadFile(platformsPath)
	if err != nil {
		return fmt.Errorf("could not read platforms file: %w", err)
	}

	// Parse YAML
	var customPlatforms map[string]*Platform
	if err := yaml.Unmarshal(data, &customPlatforms); err != nil {
		return fmt.Errorf("could not parse platforms file: %w", err)
	}

	// Add custom platforms to registry
	for id, platform := range customPlatforms {
		platform.ID = id
		platform.Custom = true
		r.Platforms[id] = platform
	}

	return nil
}

// GetPlatform returns a platform by ID
func (r *Registry) GetPlatform(id string) (*Platform, error) {
	platform, exists := r.Platforms[id]
	if !exists {
		return nil, fmt.Errorf("unknown platform: %s", id)
	}
	return platform, nil
}

// GetPlatformByHost returns a platform by host
func (r *Registry) GetPlatformByHost(host string) (*Platform, error) {
	for _, platform := range r.Platforms {
		if platform.DefaultHost == host {
			return platform, nil
		}
	}
	return nil, fmt.Errorf("unknown host: %s", host)
}

// ListPlatforms returns a list of all registered platforms
func (r *Registry) ListPlatforms() []*Platform {
	var platforms []*Platform
	for _, platform := range r.Platforms {
		platforms = append(platforms, platform)
	}
	return platforms
}

// GetProfileSSHHost returns the SSH host alias for a profile on a platform
func GetProfileSSHHost(platformID, profileName string) string {
	return fmt.Sprintf("%s-%s", platformID, profileName)
}

// GetHostAndPath extracts host and path from a URL
func GetHostAndPath(url string) (string, string, error) {
	// HTTPS format: https://github.com/user/repo.git
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
		parts := strings.SplitN(url, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid HTTPS URL format: %s", url)
		}
		return parts[0], parts[1], nil
	}

	// SSH format: git@github.com:user/repo.git
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		hostPart := strings.Split(url, "@")[1]
		hostPart = strings.Split(hostPart, ":")[0]
		pathPart := strings.Split(url, ":")[1]
		return hostPart, pathPart, nil
	}

	return "", "", fmt.Errorf("unsupported URL format: %s", url)
}

// GenerateSSHURL generates an SSH URL for the given platform, profile and path
func GenerateSSHURL(platform *Platform, profileName, path string) string {
	// Create the host alias for this platform+profile combination
	hostAlias := GetProfileSSHHost(platform.ID, profileName)

	// Return the SSH URL with the host alias
	return fmt.Sprintf("git@%s:%s", hostAlias, path)
}

// GenerateHTTPSURL generates an HTTPS URL for the given platform and path
func GenerateHTTPSURL(platform *Platform, path string) string {
	return fmt.Sprintf("https://%s/%s", platform.DefaultHost, path)
}
