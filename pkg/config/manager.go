package config

// Manager handles configuration operations
type Manager struct {
	configDir string
	config    *Config
}

// NewManager creates a new Manager with the specified config directory
func NewManager(configDir string) *Manager {
	return &Manager{
		configDir: configDir,
	}
}

// GetProfiles returns all profiles and the current profile name
func (m *Manager) GetProfiles() (map[string]Profile, string, error) {
	// Load config, handle errors, ignore validation errors for now in Manager
	validConfig, _, ioErr := LoadConfig()
	if ioErr != nil {
		return nil, "", ioErr
	}
	m.config = &validConfig // Assign address of validConfig

	return validConfig.Profiles, validConfig.Current, nil
}

// GetCurrent returns the name of the current active profile
func (m *Manager) GetCurrent() string {
	if m.config == nil {
		// Load config if not already loaded
		// Handle errors, ignore validation errors for now in Manager
		validConfig, _, ioErr := LoadConfig()
		if ioErr != nil {
			// Cannot return error here, return empty string
			return ""
		}
		m.config = &validConfig // Assign address of validConfig
	}

	return m.config.Current
}

// AddProfile adds a new profile
func (m *Manager) AddProfile(name string, profile Profile, overwrite bool) error {
	if m.config == nil {
		// Load config if not already loaded
		// Handle errors, ignore validation errors for now in Manager
		validConfig, _, ioErr := LoadConfig()
		if ioErr != nil {
			return ioErr
		}
		m.config = &validConfig // Assign address of validConfig
	}

	// Pass m.config (which is now *Config) directly
	if err := AddProfile(m.config, name, profile, overwrite); err != nil {
		return err
	}

	// Pass m.config (which is now *Config) directly
	return SaveConfig(m.config)
}

// SwitchToProfile switches to the specified profile
func (m *Manager) SwitchToProfile(name string) error {
	if m.config == nil {
		// Load config if not already loaded
		// Handle errors, ignore validation errors for now in Manager
		validConfig, _, ioErr := LoadConfig()
		if ioErr != nil {
			return ioErr
		}
		m.config = &validConfig // Assign address of validConfig
	}

	// Pass m.config (which is now *Config) directly
	if err := SwitchProfile(m.config, name); err != nil {
		return err
	}

	// Pass m.config (which is now *Config) directly
	return SaveConfig(m.config)
}

// RemoveProfile removes a profile
func (m *Manager) RemoveProfile(name string, noBackup bool) error {
	if m.config == nil {
		// Load config if not already loaded
		// Handle errors, ignore validation errors for now in Manager
		validConfig, _, ioErr := LoadConfig()
		if ioErr != nil {
			return ioErr
		}
		m.config = &validConfig // Assign address of validConfig
	}

	// Pass m.config (which is now *Config) directly
	if err := RemoveProfile(m.config, name, noBackup); err != nil {
		return err
	}

	// Pass m.config (which is now *Config) directly
	return SaveConfig(m.config)
}
