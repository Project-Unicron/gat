package git

import (
	"gat/pkg/config"
	"gat/pkg/platform"
	"gat/pkg/ssh"
)

// Manager handles Git operations
type Manager struct {
	configManager *config.Manager
	platformReg   *platform.Registry
}

// NewManager creates a new Git manager
func NewManager(configManager *config.Manager, platformReg *platform.Registry) *Manager {
	return &Manager{
		configManager: configManager,
		platformReg:   platformReg,
	}
}

// SwitchProfile switches to a different Git profile
func (m *Manager) SwitchProfile(profileName string, useSSH bool, dryRun bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Get the profile
	profiles, _, err := m.configManager.GetProfiles()
	if err != nil {
		return nil, err
	}

	profile, exists := profiles[profileName]
	if !exists {
		return nil, ErrProfileNotFound
	}

	// Don't make changes if we're in dry run mode
	if dryRun {
		result["dry_run"] = true
		result["profile"] = profile
		result["would_change"] = true
		return result, nil
	}

	// Set up Git identity
	if err := SetIdentity(profile.Username, profile.Email); err != nil {
		return nil, err
	}

	// Update Git credentials
	if err := UpdateGitCredentials(&profile); err != nil {
		return nil, err
	}

	// Set up SSH config if needed
	if profile.SSHIdentity != "" {
		sshErr := ssh.ConfigureSSH(profile.GetPlatform(), profileName, profile.SSHIdentity)
		if sshErr != nil {
			result["ssh_error"] = sshErr.Error()
		}
	}

	// Switch the active profile
	if err := m.configManager.SwitchToProfile(profileName); err != nil {
		return nil, err
	}

	// Update remote protocol if in a Git repo and useSSH flag is set
	if IsInGitRepo() {
		if useSSH {
			if err := UpdateRemoteProtocol(true, &profile, profileName); err != nil {
				result["remote_error"] = err.Error()
			}
		}
	}

	result["success"] = true
	result["profile"] = profile

	return result, nil
}

// AddProfile adds a new Git profile
func (m *Manager) AddProfile(name string, profile config.Profile, setupSSH bool, overwrite bool) error {
	// Validate the profile
	if err := validateProfile(profile); err != nil {
		return err
	}

	// Add the profile to the config
	if err := m.configManager.AddProfile(name, profile, overwrite); err != nil {
		return err
	}

	// Set up SSH if requested
	if setupSSH && profile.SSHIdentity != "" {
		if err := ssh.ConfigureSSH(profile.GetPlatform(), name, profile.SSHIdentity); err != nil {
			return err
		}
	}

	return nil
}

// RemoveProfile removes a Git profile
func (m *Manager) RemoveProfile(name string, noBackup bool) error {
	return m.configManager.RemoveProfile(name, noBackup)
}

// GetDiagnostics returns diagnostic information about the Git configuration
func (m *Manager) GetDiagnostics() (map[string]string, error) {
	return DiagnoseGitIdentity()
}

// validateProfile validates a profile's fields
func validateProfile(profile config.Profile) error {
	if !validGitHubUsername.MatchString(profile.Username) {
		return ErrInvalidUsername
	}

	if !validEmailRegex.MatchString(profile.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// Custom error variables
var (
	ErrProfileNotFound = Err("profile not found")
	ErrInvalidUsername = Err("invalid username format")
	ErrInvalidEmail    = Err("invalid email format")
)

// Err creates a new error
func Err(msg string) error {
	return Error{msg}
}

// Error is a custom error type
type Error struct {
	Message string
}

// Error returns the error message
func (e Error) Error() string {
	return e.Message
}
