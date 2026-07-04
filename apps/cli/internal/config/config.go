// Package config manages the CLI's local credentials store
// (~/.filora/config.json), with environment overrides.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultAPIURL = "http://localhost:3000"
	dirName       = ".filora"
	fileName      = "config.json"
)

// Config is the persisted CLI state.
type Config struct {
	APIURL    string `json:"api_url"`
	Token     string `json:"token,omitempty"`      // opaque CLI token (flr_...)
	SessionID string `json:"session_id,omitempty"` // id of this CLI session (for logout)
	Email     string `json:"email,omitempty"`
}

// Load reads config from disk and applies FILORA_API_URL / FILORA_TOKEN overrides.
func Load() (*Config, error) {
	cfg := &Config{APIURL: defaultAPIURL}

	path, err := filePath()
	if err != nil {
		return nil, err
	}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	if v := os.Getenv("FILORA_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("FILORA_TOKEN"); v != "" {
		cfg.Token = v
	}
	if cfg.APIURL == "" {
		cfg.APIURL = defaultAPIURL
	}
	return cfg, nil
}

// Save writes the config to disk (0600).
func (c *Config) Save() error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Clear removes stored credentials (keeps the file absent).
func Clear() error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func filePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, dirName, fileName), nil
}
