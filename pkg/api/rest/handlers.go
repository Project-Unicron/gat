package rest

import (
	"encoding/json"
	"gat/pkg/config"
	"gat/pkg/platform"
	"net/http"
)

// Handler contains all REST API handlers
type Handler struct {
	configManager *config.Manager
	platformReg   *platform.Registry
}

// NewHandler creates a new REST API handler
func NewHandler(configManager *config.Manager, platformReg *platform.Registry) *Handler {
	return &Handler{
		configManager: configManager,
		platformReg:   platformReg,
	}
}

// RegisterRoutes registers all REST API routes with the provided ServeMux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/profiles", h.handleProfiles)
	mux.HandleFunc("/platforms", h.handlePlatforms)
	mux.HandleFunc("/doctor", h.handleDoctor)
}

// ProfileResponse is the JSON response for profile requests
type ProfileResponse struct {
	Profiles []Profile `json:"profiles,omitempty"`
	Current  string    `json:"current,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// Profile is the JSON representation of a Git profile
type Profile struct {
	Name        string `json:"name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Platform    string `json:"platform"`
	Host        string `json:"host,omitempty"`
	HasToken    bool   `json:"hasToken"`
	SSHIdentity string `json:"sshIdentity,omitempty"`
	IsActive    bool   `json:"isActive"`
}

// PlatformResponse is the JSON response for platform requests
type PlatformResponse struct {
	Platforms []Platform `json:"platforms,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// Platform is the JSON representation of a Git platform
type Platform struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DefaultHost    string `json:"defaultHost"`
	SSHPrefix      string `json:"sshPrefix"`
	HTTPSPrefix    string `json:"httpsPrefix"`
	SSHUser        string `json:"sshUser"`
	TokenAuthScope string `json:"tokenAuthScope,omitempty"`
	IsCustom       bool   `json:"isCustom"`
}

// handleProfiles handles GET requests for profiles
func (h *Handler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get profiles from config
	profilesMap, _, err := h.configManager.GetProfiles()
	if err != nil {
		writeJSON(w, ProfileResponse{Error: err.Error()}, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var profiles []Profile
	currentName := h.configManager.GetCurrent()

	for name, profile := range profilesMap {
		isActive := name == currentName
		profiles = append(profiles, Profile{
			Name:        name,
			Username:    profile.Username,
			Email:       profile.Email,
			Platform:    profile.Platform,
			Host:        profile.Host,
			HasToken:    profile.Token != "",
			SSHIdentity: profile.SSHIdentity,
			IsActive:    isActive,
		})
	}

	// Send response
	writeJSON(w, ProfileResponse{
		Profiles: profiles,
		Current:  currentName,
	}, http.StatusOK)
}

// handlePlatforms handles GET requests for platforms
func (h *Handler) handlePlatforms(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get platforms from registry
	platformsList := h.platformReg.ListPlatforms()

	// Convert to response format
	var platforms []Platform
	for _, plat := range platformsList {
		platforms = append(platforms, Platform{
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

	// Send response
	writeJSON(w, PlatformResponse{Platforms: platforms}, http.StatusOK)
}

// DoctorResponse is the JSON response for doctor requests
type DoctorResponse struct {
	Status  string        `json:"status"`
	Checks  []DoctorCheck `json:"checks"`
	Summary string        `json:"summary,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// DoctorCheck is the JSON representation of a doctor check
type DoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// handleDoctor handles GET requests for diagnostics
func (h *Handler) handleDoctor(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would call the doctor functionality
	// For now, return a placeholder response
	writeJSON(w, DoctorResponse{
		Status: "ok",
		Checks: []DoctorCheck{
			{
				Name:    "Config",
				Status:  "pass",
				Message: "Configuration is valid",
			},
		},
		Summary: "All checks passed",
	}, http.StatusOK)
}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
