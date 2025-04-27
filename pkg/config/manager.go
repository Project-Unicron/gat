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
	config, err := LoadConfig()
	if err != nil {
		return nil, "", err
	}
	m.config = config

	return config.Profiles, config.Current, nil
}

// GetCurrent returns the name of the current active profile
func (m *Manager) GetCurrent() string {
	if m.config == nil {
		// Load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			return ""
		}
		m.config = config
	}

	return m.config.Current
}

// AddProfile adds a new profile
func (m *Manager) AddProfile(name string, profile Profile, overwrite bool) error {
	if m.config == nil {
		// Load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		m.config = config
	}

	if err := AddProfile(m.config, name, profile, overwrite); err != nil {
		return err
	}

	return SaveConfig(m.config)
}

// SwitchToProfile switches to the specified profile
func (m *Manager) SwitchToProfile(name string) error {
	if m.config == nil {
		// Load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		m.config = config
	}

	if err := SwitchProfile(m.config, name); err != nil {
		return err
	}

	return SaveConfig(m.config)
}

// RemoveProfile removes a profile
func (m *Manager) RemoveProfile(name string, noBackup bool) error {
	if m.config == nil {
		// Load config if not already loaded
		config, err := LoadConfig()
		if err != nil {
			return err
		}
		m.config = config
	}

	if err := RemoveProfile(m.config, name, noBackup); err != nil {
		return err
	}

	return SaveConfig(m.config)
}
