package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Validate GitHub username format - moved from pkg/git
var ValidGitHubUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]$|^[a-zA-Z0-9][a-zA-Z0-9]{0,37}$|^[a-zA-Z0-9][a-zA-Z0-9-]{0,37}[a-zA-Z0-9]$`)

// Validate Git email format
var ValidEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Profile represents a Git identity with its associated credentials
type Profile struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Token       string `json:"token,omitempty"` // Encrypted token when saved to file
	SSHIdentity string `json:"ssh_identity,omitempty"`
	Platform    string `json:"platform,omitempty"` // Platform ID (e.g., "github", "gitlab")
	Host        string `json:"host,omitempty"`     // Custom hostname if different from platform default
	AuthMethod  string `json:"auth_method"`        // Preferred authentication method ("ssh" or "https")

	// Internal fields not serialized to JSON
	rawToken string `json:"-"` // Raw, decrypted token for in-memory use
}

// Config represents the structure of the gat configuration file
type Config struct {
	Current  string             `json:"current"`
	Profiles map[string]Profile `json:"profiles"`

	// Security settings
	StoreEncrypted bool   `json:"store_encrypted"` // Whether to encrypt tokens
	NoStoreTokens  bool   `json:"no_store_tokens"` // Whether to not store tokens at all
	Salt           string `json:"salt,omitempty"`  // Salt for encryption
}

// GetToken returns the decrypted token from a profile
func (p *Profile) GetToken() string {
	if p.rawToken != "" {
		return p.rawToken
	}
	return p.Token
}

// SetToken sets the token and handles encryption if needed
func (p *Profile) SetToken(token string, encrypt bool, salt string) {
	p.rawToken = token
	if encrypt && token != "" {
		p.Token = EncryptToken(token, salt)
	} else {
		p.Token = token
	}
}

// GetPlatform returns the platform for this profile, defaulting to "github" for backwards compatibility
func (p *Profile) GetPlatform() string {
	if p.Platform == "" {
		return "github"
	}
	return p.Platform
}

// GetHost returns the host for this profile, defaulting to the platform's default host
func (p *Profile) GetHost() string {
	if p.Host != "" {
		return p.Host
	}
	return "" // Will be resolved by platform registry
}

// ConfigPath returns the path to the configuration directory
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("❌ could not find home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gat"), nil
}

// ConfigFilePath returns the path to the credentials file.
// It checks the GAT_CONFIG_FILE environment variable first.
func ConfigFilePath() (string, error) {
	// Check environment variable override
	envPath := os.Getenv("GAT_CONFIG_FILE")
	if envPath != "" {
		// Ensure the directory exists if an override is used
		envDir := filepath.Dir(envPath)
		if err := os.MkdirAll(envDir, 0700); err != nil {
			return "", fmt.Errorf("❌ could not create directory for GAT_CONFIG_FILE: %w", err)
		}
		return envPath, nil
	}

	// Default path
	configDir, err := ConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "creds.json"), nil
}

// LoadConfig loads the configuration file from disk, validates profiles,
// and returns a Config containing only valid profiles, a map of validation
// errors for invalid profiles, and any file I/O or parsing errors.
func LoadConfig() (Config, map[string]error, error) {
	configPath, err := ConfigFilePath()
	if err != nil {
		return Config{}, nil, err
	}

	// Initialize map for validation errors
	validationErrors := make(map[string]error)
	emptyValidConfig := Config{ // Used for early returns
		Profiles: make(map[string]Profile),
	}

	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return emptyValidConfig, nil, fmt.Errorf("❌ could not create config directory: %w", err)
		}

		// Create an empty config with default security settings
		emptyConfig := Config{
			Current:        "",
			Profiles:       make(map[string]Profile),
			StoreEncrypted: true,  // Default to encrypted storage
			NoStoreTokens:  false, // Store tokens by default
			Salt:           GenerateSalt(),
		}

		// Save the empty config to disk
		if err := SaveConfig(&emptyConfig); err != nil {
			return emptyValidConfig, nil, fmt.Errorf("❌ could not create initial config file: %w", err)
		}

		// Return the newly created empty config (no profiles, no errors)
		return emptyConfig, validationErrors, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return emptyValidConfig, nil, fmt.Errorf("❌ could not read config file: %w", err)
	}

	var loadedConfig Config // Holds the raw loaded config, possibly with invalid profiles
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		return emptyValidConfig, nil, fmt.Errorf("❌ could not parse config file: %w", err)
	}

	// If this is an old config file, initialize security settings
	if loadedConfig.Salt == "" {
		loadedConfig.Salt = GenerateSalt()
		loadedConfig.StoreEncrypted = true
		// Note: SaveConfig will handle persistence of these on next save
	}

	// Attempt to decrypt any tokens if they're stored encrypted
	if loadedConfig.StoreEncrypted {
		for name, profile := range loadedConfig.Profiles {
			if profile.Token != "" && strings.HasPrefix(profile.Token, "enc:") {
				decryptedToken, err := DecryptToken(profile.Token, loadedConfig.Salt)
				if err == nil {
					profile.rawToken = decryptedToken
					// Update profile in the original loaded map temporarily for validation
					loadedConfig.Profiles[name] = profile
				} else {
					// Keep encrypted token, but log a warning? Or add to validation errors?
					// For now, let's add a general decryption error, maybe not profile specific yet
					// Or perhaps mark the profile as invalid due to decryption failure?
					// Let's add it to validation errors for the specific profile.
					validationErrors[name] = fmt.Errorf("failed to decrypt token: %w", err)
					// No need to continue, validation loop later will skip this profile
				}
			}
		}
	}

	// Check and fix permissions
	EnsureSecurePermissions(configPath) // Best effort

	// Prepare the config that will hold only valid profiles
	validConfig := Config{
		Current:        loadedConfig.Current,
		Profiles:       make(map[string]Profile),
		StoreEncrypted: loadedConfig.StoreEncrypted,
		NoStoreTokens:  loadedConfig.NoStoreTokens,
		Salt:           loadedConfig.Salt,
	}

	// Validate profiles after loading
profileLoop:
	for name, profile := range loadedConfig.Profiles {
		// If decryption failed earlier, this profile is already marked invalid.
		if _, exists := validationErrors[name]; exists {
			continue profileLoop
		}

		// Validate Username
		if !ValidGitHubUsernameRegex.MatchString(profile.Username) {
			validationErrors[name] = fmt.Errorf("❌ invalid username format: '%s'", profile.Username)
			continue profileLoop
		}

		// Validate Email
		if !ValidEmailRegex.MatchString(profile.Email) {
			// Warn instead of error for email, as Git itself allows weird emails sometimes
			fmt.Printf(color.YellowString("⚠️ Warning: Profile [%s] has potentially invalid email format: %s\n"), name, profile.Email)
		}

		// Validate AuthMethod
		if profile.AuthMethod == "" {
			validationErrors[name] = fmt.Errorf("❌ missing required field 'auth_method'. Please reconfigure profile")
			continue profileLoop
		}
		profile.AuthMethod = strings.ToLower(profile.AuthMethod) // Normalize
		if profile.AuthMethod != "ssh" && profile.AuthMethod != "https" {
			validationErrors[name] = fmt.Errorf("❌ invalid auth_method: '%s'. Must be 'ssh' or 'https'", profile.AuthMethod)
			continue profileLoop
		}

		// Normalize Platform (handle legacy empty platform)
		if profile.Platform == "" {
			profile.Platform = "github" // Default for backwards compatibility
		}
		profile.Platform = strings.ToLower(profile.Platform)

		// If all checks passed, add the profile (with potentially updated fields) to the valid map
		validConfig.Profiles[name] = profile
	}

	// Ensure Current profile is actually valid, if not, unset it.
	if _, exists := validConfig.Profiles[validConfig.Current]; !exists && validConfig.Current != "" {
		// If the current profile is listed but failed validation
		if _, invalid := validationErrors[validConfig.Current]; invalid {
			fmt.Printf(color.YellowString("⚠️ Warning: Current profile [%s] is invalid, unsetting active profile.\n"), validConfig.Current)
			validConfig.Current = ""
			// Optionally, save the config here to persist the unset current profile? Or let next command handle it.
		} else {
			// This case shouldn't happen if logic is correct (current not in valid map and not in error map)
			// Maybe it was deleted manually?
			fmt.Printf(color.YellowString("⚠️ Warning: Current profile [%s] not found, unsetting active profile.\n"), validConfig.Current)
			validConfig.Current = ""
		}
	}

	return validConfig, validationErrors, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("❌ could not create config directory: %w", err)
	}

	// Handle token storage policy before saving
	processedConfig := *config

	// Process profiles for encryption or removal of tokens
	for name, profile := range processedConfig.Profiles {
		if profile.rawToken != "" {
			if config.NoStoreTokens {
				// Don't store token at all
				profile.Token = ""
			} else if config.StoreEncrypted {
				// Encrypt token before storage
				profile.Token = EncryptToken(profile.rawToken, config.Salt)
			} else {
				// Store in plaintext (with warning)
				profile.Token = profile.rawToken
			}

			// Update the profile
			processedConfig.Profiles[name] = profile
		}
	}

	data, err := json.MarshalIndent(processedConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("❌ could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("❌ could not write config file: %w", err)
	}

	// Ensure proper permissions
	if err := EnsureSecurePermissions(configPath); err != nil {
		return fmt.Errorf("❌ could not set secure permissions: %w", err)
	}

	return nil
}

// EnsureSecurePermissions ensures that config files have appropriately restrictive permissions
func EnsureSecurePermissions(path string) error {
	// On Unix-like systems, set appropriate permissions
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("could not set permissions for %s: %w", path, err)
	}

	// Also check the parent directory
	parentDir := filepath.Dir(path)
	if err := os.Chmod(parentDir, 0700); err != nil {
		return fmt.Errorf("could not set permissions for %s: %w", parentDir, err)
	}

	return nil
}

// GetCurrentProfile returns the currently active profile
func GetCurrentProfile(config *Config) (*Profile, string, error) {
	if config.Current == "" {
		return nil, "", fmt.Errorf("❌ no profile is currently active")
	}

	profile, exists := config.Profiles[config.Current]
	if !exists {
		return nil, "", fmt.Errorf("❌ active profile '%s' not found", config.Current)
	}

	return &profile, config.Current, nil
}

// AddProfile adds a new profile to the configuration
// Note: Assumes config passed in contains only valid profiles (as returned by LoadConfig)
func AddProfile(config *Config, name string, profile Profile, overwrite bool) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}

	if _, exists := config.Profiles[name]; exists && !overwrite {
		return fmt.Errorf("❌ profile [%s] already exists. Use --overwrite to replace it", name)
	}

	// Basic validation before adding (more thorough validation happens on load)
	if !ValidGitHubUsernameRegex.MatchString(profile.Username) {
		return fmt.Errorf("❌ invalid username format: '%s'", profile.Username)
	}
	if !ValidEmailRegex.MatchString(profile.Email) {
		// Allow potentially invalid emails but warn
		fmt.Printf(color.YellowString("⚠️ Warning: Profile [%s] has potentially invalid email format: %s\n"), name, profile.Email)
	}
	if profile.AuthMethod == "" {
		return fmt.Errorf("❌ 'auth_method' is required")
	}
	profile.AuthMethod = strings.ToLower(profile.AuthMethod)
	if profile.AuthMethod != "ssh" && profile.AuthMethod != "https" {
		return fmt.Errorf("❌ invalid 'auth_method': '%s'. Must be 'ssh' or 'https'", profile.AuthMethod)
	}
	if profile.Platform == "" {
		profile.Platform = "github"
	}
	profile.Platform = strings.ToLower(profile.Platform)

	config.Profiles[name] = profile
	return nil
}

// RemoveProfile removes a profile from the configuration
// Note: Assumes config passed in contains only valid profiles (as returned by LoadConfig)
func RemoveProfile(config *Config, name string, noBackup bool) error {
	if _, exists := config.Profiles[name]; !exists {
		return fmt.Errorf("❌ profile '%s' does not exist", name)
	}

	// Create backup before removal (unless explicitly disabled)
	if !noBackup {
		if err := BackupProfile(config, name); err != nil {
			return fmt.Errorf("❌ could not create backup of profile: %w", err)
		}
	}

	delete(config.Profiles, name)

	// If we deleted the current profile, unset it
	if config.Current == name {
		config.Current = ""
	}

	return nil
}

// BackupProfile creates a backup of a profile before deletion
func BackupProfile(config *Config, name string) error {
	// Create backup directory if it doesn't exist
	configDir, err := ConfigPath()
	if err != nil {
		return err
	}

	backupDir := filepath.Join(configDir, "backups")
	if err := os.MkdirAll(backupDir, 0700); err != nil {
		return fmt.Errorf("could not create backup directory: %w", err)
	}

	// Get the profile to backup
	profile, exists := config.Profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Create a backup file with timestamp
	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s.backup.json", name))

	// Create single-profile backup
	backup := map[string]Profile{
		name: profile,
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal backup: %w", err)
	}

	if err := os.WriteFile(backupFile, data, 0600); err != nil {
		return fmt.Errorf("could not write backup file: %w", err)
	}

	return nil
}

// SwitchProfile sets the current active profile
// Note: Assumes config passed in contains only valid profiles (as returned by LoadConfig)
func SwitchProfile(config *Config, name string) error {
	if _, exists := config.Profiles[name]; !exists {
		return fmt.Errorf("❌ profile [%s] not found", name)
	}

	config.Current = name
	return nil
}

// ValidateProfileName checks if a profile name is valid
// (Basic check, more comprehensive validation can be added if needed)
func ValidateProfileName(name string) error {
	// Check for empty name
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for shell special characters
	dangerousChars := []string{";", "&", "|", ">", "<", "`", "$", "\\", "\"", "'", " "}
	for _, char := range dangerousChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("profile name contains invalid character: '%s'", char)
		}
	}

	// Only allow alphanumeric, underscore, dash, and period
	validPattern := "^[a-zA-Z0-9_.-]+$"
	matched, _ := regexp.MatchString(validPattern, name)
	if !matched {
		return fmt.Errorf("profile name must contain only letters, numbers, underscore, dash, or period")
	}

	return nil
}

// EncryptToken encrypts a token using AES-256
func EncryptToken(token, salt string) string {
	if token == "" {
		return ""
	}

	// Generate key from salt
	key := deriveKey(salt)

	// Create a new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		// Fallback to plaintext on error
		return token
	}

	// Create a GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return token
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return token
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)

	// Return as base64
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext)
}

// DecryptToken decrypts a token
func DecryptToken(encryptedToken, salt string) (string, error) {
	if !strings.HasPrefix(encryptedToken, "enc:") {
		return encryptedToken, nil
	}

	// Remove prefix
	data := strings.TrimPrefix(encryptedToken, "enc:")

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	// Generate key from salt
	key := deriveKey(salt)

	// Create a new cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Split nonce and ciphertext
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateSalt generates a random salt
func GenerateSalt() string {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		// If we can't generate random data, use a timestamp
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.StdEncoding.EncodeToString(salt)
}

// deriveKey derives a cryptographic key from a salt
func deriveKey(salt string) []byte {
	hash := sha256.Sum256([]byte(salt))
	return hash[:]
}
