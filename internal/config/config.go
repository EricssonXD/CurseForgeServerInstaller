package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// configDir returns the mcserver config directory path.
func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mcserver")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "mcserver")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(configDir(), "config.json")
}

// configJSON is the on-disk JSON schema.
type configJSON struct {
	CurseForgeAPIKey string `json:"curseforgeApiKey"`
}

// AppConfig holds application configuration.
type AppConfig struct {
	CurseForgeAPIKey string
}

// Load reads the config from disk. Returns a zero-value config if the file doesn't exist.
func Load() (*AppConfig, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var raw configJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &AppConfig{
		CurseForgeAPIKey: raw.CurseForgeAPIKey,
	}, nil
}

// Save writes the config to disk with mode 0600.
func (c *AppConfig) Save() error {
	path := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	raw := configJSON{CurseForgeAPIKey: c.CurseForgeAPIKey}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// MaskSecret returns a masked version of a secret string for display.
func MaskSecret(value string) string {
	if value == "" {
		return "(not set)"
	}
	if len(value) <= 6 {
		return "***"
	}
	return value[:2] + "***" + value[len(value)-2:]
}
