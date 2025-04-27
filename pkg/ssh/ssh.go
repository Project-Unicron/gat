package ssh

import (
	"fmt"
	"gat/pkg/platform"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const gatIncludeLine = "Include ~/.ssh/gat_config"
const gatConfigComment = "# Added by gat for identity management"

// UpdateSSHConfig updates the SSH config files to manage Git host identities
func UpdateSSHConfig(platformID, profileName, sshIdentity string) error {
	if sshIdentity == "" {
		return nil // Skip if no SSH identity provided
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ could not find home directory: %w", err)
	}

	// Path to SSH directory and config files
	sshDir := filepath.Join(homeDir, ".ssh")
	mainConfigPath := filepath.Join(sshDir, "config")
	gatConfigPath := filepath.Join(sshDir, "gat_config")

	// Ensure SSH directory exists
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("❌ could not create SSH directory: %w", err)
	}

	// 1. Check if main SSH config exists, create if needed
	_, err = os.Stat(mainConfigPath)
	if os.IsNotExist(err) {
		// Create empty config file with include line
		content := fmt.Sprintf("%s\n%s\n", gatConfigComment, gatIncludeLine)
		if err := os.WriteFile(mainConfigPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("❌ could not create SSH config: %w", err)
		}
		fmt.Println("🔐 Created SSH config file with gat include")
	} else if err == nil {
		// Check if the include line exists, add if missing
		if err := ensureGatIncludeLine(mainConfigPath); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("❌ could not check SSH config: %w", err)
	}

	// 2. Update the gat_config file with the platform-specific host
	if err := updateGatConfig(gatConfigPath, platformID, profileName, sshIdentity); err != nil {
		return err
	}

	return nil
}

// ensureGatIncludeLine checks if the gat include line exists in the SSH config
// and adds it if it's missing
func ensureGatIncludeLine(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("❌ could not read SSH config: %w", err)
	}

	content := string(data)
	includePattern := regexp.MustCompile(fmt.Sprintf(`(?m)^%s$`, regexp.QuoteMeta(gatIncludeLine)))

	if !includePattern.MatchString(content) {
		// Add the include line at the end
		newContent := content
		if !strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		}
		newContent += "\n" + gatConfigComment + "\n" + gatIncludeLine + "\n"

		if err := os.WriteFile(configPath, []byte(newContent), 0600); err != nil {
			return fmt.Errorf("❌ could not update SSH config: %w", err)
		}
		fmt.Println("🔐 Updated SSH config to include gat configuration")
	}

	return nil
}

// updateGatConfig updates the gat_config file with the platform-specific host
func updateGatConfig(configPath, platformID, profileName, sshIdentity string) error {
	// Format the identity path based on platform
	formattedIdentity := formatSSHPath(sshIdentity)

	// Generate host alias for this platform+profile combination
	hostAlias := platform.GetProfileSSHHost(platformID, profileName)

	// Get platform info from registry (default to github.com if not found)
	reg := platform.NewRegistry()
	plat, err := reg.GetPlatform(platformID)
	if err != nil {
		plat = &platform.Platform{
			DefaultHost: "github.com",
			SSHUser:     "git",
		}
	}

	// Define the host block template
	hostBlock := fmt.Sprintf(`
# Profile: %s on %s (managed by gat)
Host %s
    HostName %s
    User %s
    IdentityFile %s
    IdentitiesOnly yes
`, profileName, plat.Name, hostAlias, plat.DefaultHost, plat.SSHUser, formattedIdentity)

	// Check if the file exists
	data, err := os.ReadFile(configPath)

	var content string
	if os.IsNotExist(err) {
		// Create new file with the host block
		content = hostBlock
	} else if err != nil {
		return fmt.Errorf("❌ could not read gat SSH config: %w", err)
	} else {
		// File exists, update or add the host block
		content = string(data)

		// Check for existing entry for this host alias
		hostPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^Host %s$`, regexp.QuoteMeta(hostAlias)))
		if hostPattern.MatchString(content) {
			// Replace existing block
			profilePattern := regexp.MustCompile(fmt.Sprintf(`(?ms)# Profile:.*?Host %s.*?(^\s*$|^Host)`,
				regexp.QuoteMeta(hostAlias)))

			if profilePattern.MatchString(content) {
				content = profilePattern.ReplaceAllString(content, hostBlock+"\n")
			} else {
				// If pattern doesn't match exactly, remove the Host line and append a new block
				content = hostPattern.ReplaceAllString(content, "") // Remove the Host line
				if !strings.HasSuffix(content, "\n") {
					content += "\n"
				}
				content += hostBlock
			}
		} else {
			// Append new block
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += hostBlock
		}
	}

	// Write the updated content
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("❌ could not write gat SSH config: %w", err)
	}

	fmt.Printf("🔐 Updated SSH configuration for %s profile: %s\n", platformID, profileName)
	return nil
}

// formatSSHPath formats the SSH identity path based on the current platform
func formatSSHPath(sshIdentity string) string {
	// On Windows, convert backslashes to forward slashes in the SSH config
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(sshIdentity, "\\", "/")
	}
	return sshIdentity
}

// GetSSHCloneURL returns an SSH clone URL for a specific profile
func GetSSHCloneURL(repoURL, platformID, profileName string) string {
	// Extract repository path from URL
	repoPath := extractRepoPath(repoURL)

	// Create host alias for this platform+profile
	hostAlias := platform.GetProfileSSHHost(platformID, profileName)

	return fmt.Sprintf("git@%s:%s", hostAlias, repoPath)
}

// extractRepoPath extracts the repository path from a URL
func extractRepoPath(repoURL string) string {
	// Extract host and path
	_, path, err := platform.GetHostAndPath(repoURL)
	if err != nil {
		return strings.TrimSuffix(repoURL, ".git")
	}

	// Remove .git suffix if present
	return strings.TrimSuffix(path, ".git")
}

// CheckSSHIdentity checks if an SSH identity file exists
func CheckSSHIdentity(sshIdentity string) (bool, error) {
	if sshIdentity == "" {
		return false, nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(sshIdentity, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false, fmt.Errorf("❌ could not find home directory: %w", err)
		}
		sshIdentity = filepath.Join(homeDir, sshIdentity[1:])
	}

	// Check if identity file exists
	_, err := os.Stat(sshIdentity)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("❌ could not check SSH identity: %w", err)
	}

	// Also check if the .pub file exists
	_, err = os.Stat(sshIdentity + ".pub")
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("❌ could not check SSH public key: %w", err)
	}

	return true, nil
}

// CheckSSHSetup checks if the SSH configuration is set up correctly for gat
func CheckSSHSetup() (bool, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("❌ could not find home directory: %w", err)
	}

	// Path to SSH config files
	mainConfigPath := filepath.Join(homeDir, ".ssh", "config")
	gatConfigPath := filepath.Join(homeDir, ".ssh", "gat_config")

	// Check if main SSH config exists
	_, err = os.Stat(mainConfigPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("❌ could not check SSH config: %w", err)
	}

	// Check if main config includes gat_config
	data, err := os.ReadFile(mainConfigPath)
	if err != nil {
		return false, fmt.Errorf("❌ could not read SSH config: %w", err)
	}

	includePattern := regexp.MustCompile(fmt.Sprintf(`(?m)^%s$`, regexp.QuoteMeta(gatIncludeLine)))
	if !includePattern.MatchString(string(data)) {
		return false, nil
	}

	// Check if gat_config exists
	_, err = os.Stat(gatConfigPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("❌ could not check gat SSH config: %w", err)
	}

	return true, nil
}

// ConfigureSSH configures SSH for a specific profile
func ConfigureSSH(platformID, profileName, sshIdentity string) error {
	// Get config path
	configPath, err := getGatConfigPath()
	if err != nil {
		return err
	}

	// Update GAT-specific SSH config
	return updateGatConfig(configPath, platformID, profileName, sshIdentity)
}

// getGatConfigPath returns the path to the gat SSH config file
func getGatConfigPath() (string, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("❌ could not find home directory: %w", err)
	}

	// Return path to gat_config file in .ssh directory
	return filepath.Join(homeDir, ".ssh", "gat_config"), nil
}
