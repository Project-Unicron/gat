package ssh

import (
	"fmt"
	"gat/pkg/platform"
	"os"
	"os/exec"
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

// StartAgent ensures the ssh-agent is running.
// Returns an error if it cannot start or connect to the agent.
func StartAgent() error {
	// Check if agent is already running by checking environment variable
	if os.Getenv("SSH_AUTH_SOCK") != "" {
		// Agent seems to be running, try listing keys to confirm connection
		cmd := exec.Command("ssh-add", "-l")
		if err := cmd.Run(); err == nil {
			return nil // Agent is running and accessible
		}
	}

	// Agent not running or not accessible, try starting it
	fmt.Println("🔑 Starting ssh-agent...")
	cmd := exec.Command("ssh-agent", "-s")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("❌ failed to start ssh-agent: %w\nOutput: %s", err, string(output))
	}

	// Parse the output to set environment variables (SSH_AUTH_SOCK, SSH_AGENT_PID)
	// Example output:
	// SSH_AUTH_SOCK=/tmp/ssh-XXXXXXXXXX/agent.pid; export SSH_AUTH_SOCK;
	// SSH_AGENT_PID=12345; export SSH_AGENT_PID;
	// echo Agent pid 12345;
	lines := strings.Split(string(output), ";")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSH_AUTH_SOCK=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				os.Setenv("SSH_AUTH_SOCK", parts[1])
			}
		} else if strings.HasPrefix(line, "SSH_AGENT_PID=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				os.Setenv("SSH_AGENT_PID", parts[1])
			}
		}
	}

	// Verify agent started by checking env var again
	if os.Getenv("SSH_AUTH_SOCK") == "" {
		return fmt.Errorf("❌ failed to parse ssh-agent output or set environment variables")
	}

	fmt.Println("✅ ssh-agent started")
	return nil
}

// ClearIdentities removes all identities from the ssh-agent.
func ClearIdentities() error {
	fmt.Println("🧹 Clearing existing SSH identities from agent...")
	cmd := exec.Command("ssh-add", "-D")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is just "Agent has no identities"
		if strings.Contains(string(output), "Agent has no identities") || strings.Contains(string(output), "Could not remove all identities") {
			fmt.Println("ℹ️ No identities to clear or agent was empty.")
			return nil // Not a fatal error
		}
		return fmt.Errorf("❌ failed to clear ssh-agent identities: %w\nOutput: %s", err, string(output))
	}
	fmt.Println("✅ Identities cleared")
	return nil
}

// AddIdentity adds a specific SSH identity to the ssh-agent.
func AddIdentity(identityPath string) error {
	fmt.Printf("➕ Adding SSH identity: %s\n", identityPath)

	// Expand ~ to home directory
	if strings.HasPrefix(identityPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("❌ could not find home directory: %w", err)
		}
		identityPath = filepath.Join(homeDir, identityPath[1:])
	}

	cmd := exec.Command("ssh-add", identityPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ failed to add SSH identity '%s': %w\nOutput: %s", identityPath, err, string(output))
	}

	// Check output for success message (ssh-add output varies)
	if !strings.Contains(string(output), "Identity added") {
		// Some versions might just output nothing on success, check error code was 0
		if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
			return fmt.Errorf("❌ unknown error adding SSH identity '%s'\nOutput: %s", identityPath, string(output))
		}
	}

	fmt.Printf("✅ Identity added: %s\n", identityPath)
	return nil
}
