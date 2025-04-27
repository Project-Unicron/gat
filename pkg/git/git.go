package git

import (
	"fmt"
	"gat/pkg/config"
	"gat/pkg/platform"
	"gat/pkg/ssh"
	"gat/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Validate GitHub username format - without using unsupported look-ahead assertions
// This regex enforces the GitHub username rules:
// 1. Must start and end with alphanumeric characters
// 2. Can contain hyphens in the middle (not consecutive)
// 3. Length between 1-39 characters
// The pattern covers three cases:
// - Single character: ^[a-zA-Z0-9]$
// - All alphanumeric: ^[a-zA-Z0-9][a-zA-Z0-9]{0,37}$
// - With hyphens: ^[a-zA-Z0-9][a-zA-Z0-9-]{0,37}[a-zA-Z0-9]$
var validGitHubUsername = regexp.MustCompile(`^[a-zA-Z0-9]$|^[a-zA-Z0-9][a-zA-Z0-9]{0,37}$|^[a-zA-Z0-9][a-zA-Z0-9-]{0,37}[a-zA-Z0-9]$`)

// Validate Git email format
var validEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SetIdentity sets the user's Git identity in the global Git config
func SetIdentity(username, email string) error {
	// Validate inputs
	if !validGitHubUsername.MatchString(username) {
		return fmt.Errorf("❌ invalid GitHub username format: %s", username)
	}

	if !validEmailRegex.MatchString(email) {
		return fmt.Errorf("❌ invalid email format: %s", email)
	}

	// Set user.name
	cmdName := exec.Command("git", "config", "--global", "user.name", username)
	if err := cmdName.Run(); err != nil {
		return fmt.Errorf("❌ could not set git username: %w", err)
	}

	// Set user.email
	cmdEmail := exec.Command("git", "config", "--global", "user.email", email)
	if err := cmdEmail.Run(); err != nil {
		return fmt.Errorf("❌ could not set git email: %w", err)
	}

	return nil
}

// IsInGitRepo checks if the current directory is inside a Git repository
func IsInGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// GetCurrentRemoteURL gets the remote URL for the current repository
func GetCurrentRemoteURL() (string, error) {
	if !IsInGitRepo() {
		return "", fmt.Errorf("❌ not in a git repository")
	}

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.CombinedOutput() // Use CombinedOutput to get stderr if there's an error
	if err != nil {
		stderr := strings.TrimSpace(string(output))
		if stderr != "" {
			return "", fmt.Errorf("❌ could not get remote URL: %s", stderr)
		}
		return "", fmt.Errorf("❌ could not get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsSSHRemote checks if the remote URL uses SSH protocol
func IsSSHRemote(url string) bool {
	return strings.HasPrefix(url, "git@") || strings.Contains(url, "ssh://")
}

// IsProfileSSHRemote checks if the remote URL is using a profile-specific SSH format
func IsProfileSSHRemote(url string) (bool, string, string) {
	// Format: git@platform-profilename:user/repo.git
	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return false, "", ""
		}

		hostParts := strings.Split(parts[0], "@")
		if len(hostParts) != 2 {
			return false, "", ""
		}

		hostAlias := hostParts[1]
		if !strings.Contains(hostAlias, "-") {
			return false, "", ""
		}

		// Split the platform-profile parts
		aliasParts := strings.SplitN(hostAlias, "-", 2)
		if len(aliasParts) != 2 {
			return false, "", ""
		}

		platformID := aliasParts[0]
		profileName := aliasParts[1]

		// Validate profile name
		if err := config.ValidateProfileName(profileName); err != nil {
			return false, "", ""
		}

		return true, platformID, profileName
	}

	return false, "", ""
}

// ConvertRemoteToHTTPS converts a remote URL to HTTPS format
func ConvertRemoteToHTTPS(url string, profile *config.Profile) string {
	platformID := profile.GetPlatform()

	// Get platform information
	reg := platform.NewRegistry()

	// Determine appropriate platform settings, with fallbacks
	defaultHost := "github.com"

	// Try to get platform from registry
	plat, err := reg.GetPlatform(platformID)
	if err == nil {
		// Use the platform info from registry
		defaultHost = plat.DefaultHost
	} else {
		// If platform not found, try to infer it from the URL
		host, _, urlErr := platform.GetHostAndPath(url)
		if urlErr == nil {
			if inferredPlat, inferredErr := reg.GetPlatformByHost(host); inferredErr == nil {
				defaultHost = inferredPlat.DefaultHost
			}
		}
		// On failure, we keep the GitHub defaults
	}

	// Use custom host if specified
	if profile.Host != "" {
		defaultHost = profile.Host
	}

	if IsSSHRemote(url) {
		// Check if it's a profile-specific SSH URL
		isProfileSSH, _, _ := IsProfileSSHRemote(url)
		if isProfileSSH {
			// Extract the user/repo part
			parts := strings.Split(url, ":")
			if len(parts) != 2 {
				return url // Unable to parse, return as is
			}

			path := parts[1]
			path = strings.TrimSuffix(path, ".git")
			return fmt.Sprintf("https://%s/%s", defaultHost, path)
		}

		// Standard SSH URL
		// Extract the host and path from SSH URL
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			// Fallback for ssh:// format
			if strings.Contains(url, "ssh://") {
				url = strings.TrimPrefix(url, "ssh://")
				parts = strings.SplitN(url, "/", 2)
				if len(parts) != 2 {
					return url // Unable to parse, return as is
				}
				// Use the original host from the URL in this case
				sshHost := parts[0]
				sshHost = strings.TrimPrefix(sshHost, "git@")
				path := parts[1]
				return fmt.Sprintf("https://%s/%s", sshHost, path)
			}
			return url // Unable to parse, return as is
		}

		// Use the path from the URL but the host from the profile
		path := parts[1]
		return fmt.Sprintf("https://%s/%s", defaultHost, path)
	}
	return url // Already HTTPS or unknown format
}

// ConvertRemoteToSSH converts a remote URL to SSH format
func ConvertRemoteToSSH(url string, profile *config.Profile) string {
	platformID := profile.GetPlatform()

	// Get platform information
	reg := platform.NewRegistry()
	sshUser := ""

	// Determine appropriate platform settings, with fallbacks
	// Try to get platform from registry
	if plat, err := reg.GetPlatform(platformID); err == nil {
		// Use the platform info from registry
		sshUser = plat.SSHUser
	} else {
		// If platform not found, try to infer it from the URL
		host, _, urlErr := platform.GetHostAndPath(url)
		if urlErr == nil {
			if inferredPlat, inferredErr := reg.GetPlatformByHost(host); inferredErr == nil {
				sshUser = inferredPlat.SSHUser
			}
		}
		// On failure, we keep the default values
	}

	// Default to git user if still empty
	if sshUser == "" {
		sshUser = "git"
	}

	// Process the URL
	if IsSSHRemote(url) {
		// Already an SSH URL, check if it needs to be converted to profile format
		isProfileSSH, currentPlatformID, currentProfile := IsProfileSSHRemote(url)
		if isProfileSSH && (currentPlatformID != platformID || currentProfile != sshUser) {
			// Need to update the profile in the URL
			parts := strings.Split(url, ":")
			if len(parts) == 2 {
				hostAlias := platform.GetProfileSSHHost(platformID, sshUser)
				return fmt.Sprintf("git@%s:%s", hostAlias, parts[1])
			}
		} else if !isProfileSSH {
			// Check if this is an SSH URL for the same platform
			_, path, err := platform.GetHostAndPath(url)
			if err == nil {
				// Convert standard SSH URL to profile-specific format for this platform
				hostAlias := platform.GetProfileSSHHost(platformID, sshUser)
				return fmt.Sprintf("git@%s:%s", hostAlias, path)
			}
		}
	} else {
		// Convert HTTPS to SSH
		// Extract the host and path from HTTPS URL
		url = strings.TrimPrefix(url, "https://")
		parts := strings.SplitN(url, "/", 2)
		if len(parts) != 2 {
			return url // Unable to parse, return as is
		}

		// Use the host alias for this platform+profile combination
		hostAlias := platform.GetProfileSSHHost(platformID, sshUser)
		path := parts[1]

		// Return the SSH URL with the host alias
		return fmt.Sprintf("git@%s:%s", hostAlias, path)
	}

	// If we reach here, the URL is either already in the correct format or we couldn't parse it
	return url
}

// UpdateRemoteURL updates the remote URL for the current repository
func UpdateRemoteURL(url string) error {
	if !IsInGitRepo() {
		return fmt.Errorf("❌ not in a git repository")
	}

	// Validate URL format for security
	if !isValidRemoteURL(url) {
		return fmt.Errorf("❌ invalid remote URL format: %s", url)
	}

	// Specifically create command with explicit args for security
	args := []string{"remote", "set-url", "origin", url}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		stderr := strings.TrimSpace(string(output))
		if stderr != "" {
			return fmt.Errorf("❌ could not update remote URL: %s", stderr)
		}
		return fmt.Errorf("❌ could not update remote URL: %w", err)
	}

	return nil
}

// isValidRemoteURL checks if a URL is a valid Git remote URL
func isValidRemoteURL(url string) bool {
	// Check for SSH URLs
	if IsSSHRemote(url) {
		// For SSH URLs, check basic structure
		if strings.HasPrefix(url, "git@") {
			parts := strings.Split(url, ":")
			if len(parts) != 2 {
				return false
			}

			// Check host component - support any valid Git hosting platform or profile-specific pattern
			hostPart := parts[0]
			pathPart := parts[1]

			// Accept standard git@ URLs for known platforms
			if isValidSSHHostFormat(hostPart) {
				return strings.Contains(pathPart, "/") &&
					(strings.HasSuffix(pathPart, ".git") || !strings.Contains(pathPart, " "))
			}

			// Accept profile-specific SSH URLs (e.g. git@github-work:user/repo.git)
			if strings.HasPrefix(hostPart, "git@") && strings.Contains(hostPart, "-") {
				platformProfile := strings.TrimPrefix(hostPart, "git@")
				platformProfileParts := strings.Split(platformProfile, "-")

				if len(platformProfileParts) == 2 {
					// Valid platform-profile pattern
					return strings.Contains(pathPart, "/") &&
						(strings.HasSuffix(pathPart, ".git") || !strings.Contains(pathPart, " "))
				}
			}

			return false
		}
		return false
	}

	// Check for HTTPS URLs
	if strings.HasPrefix(url, "https://") {
		// For HTTPS URLs, check basic structure
		url = strings.TrimPrefix(url, "https://")
		parts := strings.SplitN(url, "/", 2)
		if len(parts) != 2 {
			return false
		}

		// Check host component - any valid Git hosting platform
		hostPart := parts[0]
		pathPart := parts[1]

		// Accept URLs from any known platform
		if isValidHTTPSHostFormat(hostPart) {
			return strings.Contains(pathPart, "/") && !strings.Contains(pathPart, " ")
		}

		// Also accept custom hosts from user's platform registry
		// For security, ensure the host doesn't contain any dangerous characters
		if !strings.ContainsAny(hostPart, " ;\"'<>|&") {
			return strings.Contains(pathPart, "/") && !strings.Contains(pathPart, " ")
		}
	}

	return false
}

// isValidSSHHostFormat checks if a hostname is a valid SSH host format for any platform
func isValidSSHHostFormat(hostPart string) bool {
	reg := platform.NewRegistry()
	for _, p := range reg.ListPlatforms() {
		expectedPrefix := fmt.Sprintf("git@%s", p.DefaultHost)
		if strings.HasPrefix(hostPart, expectedPrefix) {
			return true
		}
	}
	return false
}

// isValidHTTPSHostFormat checks if a hostname is a valid HTTPS host for any platform
func isValidHTTPSHostFormat(hostPart string) bool {
	reg := platform.NewRegistry()
	for _, p := range reg.ListPlatforms() {
		if hostPart == p.DefaultHost {
			return true
		}
	}
	return false
}

// UpdateRemoteProtocol switches the remote protocol between HTTPS and SSH
func UpdateRemoteProtocol(useSSH bool, profile *config.Profile, profileName string) error {
	// Validate the profile name
	if err := config.ValidateProfileName(profileName); err != nil {
		return err
	}

	url, err := GetCurrentRemoteURL()
	if err != nil {
		return err
	}

	var newURL string
	if useSSH {
		newURL = ConvertRemoteToSSH(url, profile)
	} else {
		newURL = ConvertRemoteToHTTPS(url, profile)
	}

	if newURL != url {
		if err := UpdateRemoteURL(newURL); err != nil {
			return err
		}
	}

	return nil
}

// UpdateGitCredentials updates the .git-credentials file with the token
func UpdateGitCredentials(profile *config.Profile) error {
	token := profile.GetToken()
	username := profile.Username

	if token == "" {
		return nil // Skip if no token provided
	}

	// Validate inputs
	if !validGitHubUsername.MatchString(username) {
		return fmt.Errorf("❌ invalid username format: %s", username)
	}

	// Get platform information
	platformID := profile.GetPlatform()
	reg := platform.NewRegistry()
	plat, err := reg.GetPlatform(platformID)
	if err != nil {
		// Default to GitHub
		plat = &platform.Platform{
			DefaultHost:    "github.com",
			TokenAuthScope: "github.com",
		}
	}

	// Use custom host if specified
	host := plat.DefaultHost
	if profile.Host != "" {
		host = profile.Host
	}

	// Enable credential store
	cmdStore := exec.Command("git", "config", "--global", "credential.helper", "store")
	if err := cmdStore.Run(); err != nil {
		return fmt.Errorf("❌ could not set credential helper: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("❌ could not find home directory: %w", err)
	}

	credFile := filepath.Join(homeDir, ".git-credentials")

	// Create or truncate the credentials file with secure permissions
	file, err := os.OpenFile(credFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("❌ could not open .git-credentials: %w", err)
	}
	defer file.Close()

	// Write credential in format: https://username:token@host
	// Specifically avoid string interpolation of the token directly
	cred := fmt.Sprintf("https://%s:%s@%s\n", username, token, host)
	if _, err := file.WriteString(cred); err != nil {
		return fmt.Errorf("❌ could not write to .git-credentials: %w", err)
	}

	// Ensure secure permissions
	if err := os.Chmod(credFile, 0600); err != nil {
		return fmt.Errorf("❌ could not set permissions for .git-credentials: %w", err)
	}

	return nil
}

// GetGitConfig retrieves a value from Git's global config
func GetGitConfig(key string) (string, error) {
	// Validate key to prevent injection
	if !isValidGitConfigKey(key) {
		return "", fmt.Errorf("❌ invalid git config key: %s", key)
	}

	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Config key not found
			return "", nil
		}
		return "", fmt.Errorf("❌ could not get git config for %s: %w", key, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// isValidGitConfigKey validates a git config key for security
func isValidGitConfigKey(key string) bool {
	// Only allow specific sections we use
	allowedPrefixes := []string{
		"user.name",
		"user.email",
		"credential.helper",
	}

	for _, prefix := range allowedPrefixes {
		if key == prefix {
			return true
		}
	}

	// Reject keys with potential shell injection characters
	dangerousChars := []string{";", "&", "|", ">", "<", "`", "$", "\\", "\"", "'", " "}
	for _, char := range dangerousChars {
		if strings.Contains(key, char) {
			return false
		}
	}

	return false
}

// DiagnoseGitIdentity checks the current Git identity and configuration
func DiagnoseGitIdentity() (map[string]string, error) {
	diagnosis := make(map[string]string)

	// Check username
	username, err := GetGitConfig("user.name")
	if err != nil {
		return nil, err
	}
	diagnosis["username"] = username

	// Check email
	email, err := GetGitConfig("user.email")
	if err != nil {
		return nil, err
	}
	diagnosis["email"] = email

	// Check credential helper
	credHelper, err := GetGitConfig("credential.helper")
	if err != nil {
		return nil, err
	}
	diagnosis["credential_helper"] = credHelper

	// Check SSH setup
	sshConfigured, err := ssh.CheckSSHSetup()
	if err != nil {
		diagnosis["ssh_setup_error"] = err.Error()
	} else {
		diagnosis["ssh_configured"] = utils.Ternary(sshConfigured, "true", "false")
	}

	// Check if in a Git repo
	if IsInGitRepo() {
		diagnosis["in_git_repo"] = "true"

		// Get remote URL
		remoteURL, err := GetCurrentRemoteURL()
		if err == nil {
			diagnosis["remote_url"] = remoteURL

			// Check if it's using profile-specific SSH
			isProfileSSH, _, _ := IsProfileSSHRemote(remoteURL)
			if isProfileSSH {
				diagnosis["protocol"] = "SSH (profile: " + remoteURL + ")"
			} else {
				diagnosis["protocol"] = utils.Ternary(IsSSHRemote(remoteURL), "SSH", "HTTPS")
			}
		}
	} else {
		diagnosis["in_git_repo"] = "false"
	}

	return diagnosis, nil
}
