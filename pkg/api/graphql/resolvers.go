package graphql

import (
	"context"
	"gat/pkg/config"
	"gat/pkg/git"
	"gat/pkg/platform"
)

// Resolver is the root resolver for GraphQL queries
type Resolver struct {
	configManager *config.Manager
	platformReg   *platform.Registry
	gitManager    *git.Manager
}

// NewResolver creates a new root resolver
func NewResolver(configManager *config.Manager, platformReg *platform.Registry, gitManager *git.Manager) *Resolver {
	return &Resolver{
		configManager: configManager,
		platformReg:   platformReg,
		gitManager:    gitManager,
	}
}

// Profile represents a Git profile
type Profile struct {
	Name        string
	Username    string
	Email       string
	Platform    string
	Host        string
	Token       string
	SSHIdentity string
	IsActive    bool
}

// Platform represents a Git hosting platform
type Platform struct {
	ID             string
	Name           string
	DefaultHost    string
	SSHPrefix      string
	HTTPSPrefix    string
	SSHUser        string
	TokenAuthScope string
	IsCustom       bool
}

// HasToken returns whether the profile has a token
func (p *Profile) HasToken() bool {
	return p.Token != ""
}

// ProfileDetailsResolver resolves platform details for a profile
func (r *Resolver) ProfileDetailsResolver(ctx context.Context, profile *Profile) (*Platform, error) {
	platID := profile.Platform
	plat, err := r.platformReg.GetPlatform(platID)
	if err != nil {
		return nil, err
	}

	return &Platform{
		ID:             plat.ID,
		Name:           plat.Name,
		DefaultHost:    plat.DefaultHost,
		SSHPrefix:      plat.SSHPrefix,
		HTTPSPrefix:    plat.HTTPSPrefix,
		SSHUser:        plat.SSHUser,
		TokenAuthScope: plat.TokenAuthScope,
		IsCustom:       plat.Custom,
	}, nil
}

// Profiles returns all profiles
func (r *Resolver) Profiles(ctx context.Context) ([]*Profile, error) {
	profilesMap, _, err := r.configManager.GetProfiles()
	if err != nil {
		return nil, err
	}

	var profiles []*Profile
	for name, profile := range profilesMap {
		isActive := name == r.configManager.GetCurrent()
		profiles = append(profiles, &Profile{
			Name:        name,
			Username:    profile.Username,
			Email:       profile.Email,
			Platform:    profile.Platform,
			Host:        profile.Host,
			Token:       profile.Token,
			SSHIdentity: profile.SSHIdentity,
			IsActive:    isActive,
		})
	}

	return profiles, nil
}

// Profile returns a specific profile by name
func (r *Resolver) Profile(ctx context.Context, args struct{ Name string }) (*Profile, error) {
	profilesMap, _, err := r.configManager.GetProfiles()
	if err != nil {
		return nil, err
	}

	profile, exists := profilesMap[args.Name]
	if !exists {
		return nil, nil // Return nil for not found
	}

	isActive := args.Name == r.configManager.GetCurrent()
	return &Profile{
		Name:        args.Name,
		Username:    profile.Username,
		Email:       profile.Email,
		Platform:    profile.Platform,
		Host:        profile.Host,
		Token:       profile.Token,
		SSHIdentity: profile.SSHIdentity,
		IsActive:    isActive,
	}, nil
}

// CurrentProfile returns the current active profile
func (r *Resolver) CurrentProfile(ctx context.Context) (*Profile, error) {
	profilesMap, _, err := r.configManager.GetProfiles()
	if err != nil {
		return nil, err
	}

	currentName := r.configManager.GetCurrent()
	if currentName == "" {
		return nil, nil // No current profile
	}

	profile, exists := profilesMap[currentName]
	if !exists {
		return nil, nil // Should not happen, but handle anyway
	}

	return &Profile{
		Name:        currentName,
		Username:    profile.Username,
		Email:       profile.Email,
		Platform:    profile.Platform,
		Host:        profile.Host,
		Token:       profile.Token,
		SSHIdentity: profile.SSHIdentity,
		IsActive:    true,
	}, nil
}

// Platforms returns all platforms
func (r *Resolver) Platforms(ctx context.Context) ([]*Platform, error) {
	platsList := r.platformReg.ListPlatforms()

	var platforms []*Platform
	for _, plat := range platsList {
		platforms = append(platforms, &Platform{
			ID:             plat.ID,
			Name:           plat.Name,
			DefaultHost:    plat.DefaultHost,
			SSHPrefix:      plat.SSHPrefix,
			HTTPSPrefix:    plat.HTTPSPrefix,
			SSHUser:        plat.SSHUser,
			TokenAuthScope: plat.TokenAuthScope,
			IsCustom:       plat.Custom,
		})
	}

	return platforms, nil
}

// Platform returns a specific platform by ID
func (r *Resolver) Platform(ctx context.Context, args struct{ ID string }) (*Platform, error) {
	plat, err := r.platformReg.GetPlatform(args.ID)
	if err != nil {
		return nil, nil // Return nil for not found
	}

	return &Platform{
		ID:             plat.ID,
		Name:           plat.Name,
		DefaultHost:    plat.DefaultHost,
		SSHPrefix:      plat.SSHPrefix,
		HTTPSPrefix:    plat.HTTPSPrefix,
		SSHUser:        plat.SSHUser,
		TokenAuthScope: plat.TokenAuthScope,
		IsCustom:       plat.Custom,
	}, nil
}

// SwitchProfileInput represents input for switching profiles
type SwitchProfileInput struct {
	Name     string
	Protocol *string
	DryRun   *bool
}

// SwitchProfileResult represents the result of a profile switch
type SwitchProfileResult struct {
	Success   bool
	Message   *string
	Profile   *Profile
	GitConfig []*GitConfigChange
}

// GitConfigChange represents a Git config change
type GitConfigChange struct {
	Key      string
	OldValue *string
	NewValue *string
}

// SwitchProfile switches to a different profile
func (r *Resolver) SwitchProfile(ctx context.Context, args struct{ Input SwitchProfileInput }) (*SwitchProfileResult, error) {
	// Implementation would call the existing switch profile functionality
	// This is a placeholder for now
	return &SwitchProfileResult{
		Success: true,
		Message: strPtr("Profile switched successfully"),
		// Would populate Profile and GitConfig
	}, nil
}

// Helper to create string pointers
func strPtr(s string) *string {
	return &s
}
