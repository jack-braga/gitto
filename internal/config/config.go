package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// Config holds all user-configurable settings for gitto.
type Config struct {
	Editor         string `json:"editor"`
	Theme          string `json:"theme"`
	PollInterval   int    `json:"poll_interval"`
	ShowUntracked  bool   `json:"show_untracked"`
	ShowIgnored    bool   `json:"show_ignored"`
	ConfirmDiscard bool   `json:"confirm_discard"`
	ConfirmForce   bool   `json:"confirm_force"`
	DefaultView    string `json:"default_view"`
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		Editor:         "",
		Theme:          "auto",
		PollInterval:   5,
		ShowUntracked:  true,
		ShowIgnored:    false,
		ConfirmDiscard: true,
		ConfirmForce:   true,
		DefaultView:    "source",
	}
}

// configDir returns the platform-appropriate config directory.
func configDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "gitto")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "gitto")
	}

	// XDG base directory spec
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg != "" {
		return filepath.Join(xdg, "gitto")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gitto")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(configDir(), "config.json")
}

// Load reads the config from disk. If the file doesn't exist, returns defaults.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(ConfigPath(), data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Get returns a config value by key name as a string.
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "editor":
		return cfg.Editor, nil
	case "theme":
		return cfg.Theme, nil
	case "poll_interval":
		return strconv.Itoa(cfg.PollInterval), nil
	case "show_untracked":
		return strconv.FormatBool(cfg.ShowUntracked), nil
	case "show_ignored":
		return strconv.FormatBool(cfg.ShowIgnored), nil
	case "confirm_discard":
		return strconv.FormatBool(cfg.ConfirmDiscard), nil
	case "confirm_force":
		return strconv.FormatBool(cfg.ConfirmForce), nil
	case "default_view":
		return cfg.DefaultView, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Set updates a config value by key name, handling type conversions.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "editor":
		cfg.Editor = value
	case "theme":
		if value != "auto" && value != "dark" && value != "light" {
			return fmt.Errorf("theme must be auto, dark, or light")
		}
		cfg.Theme = value
	case "poll_interval":
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 {
			return fmt.Errorf("poll_interval must be a positive integer")
		}
		cfg.PollInterval = n
	case "show_untracked":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("show_untracked must be true or false")
		}
		cfg.ShowUntracked = b
	case "show_ignored":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("show_ignored must be true or false")
		}
		cfg.ShowIgnored = b
	case "confirm_discard":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("confirm_discard must be true or false")
		}
		cfg.ConfirmDiscard = b
	case "confirm_force":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("confirm_force must be true or false")
		}
		cfg.ConfirmForce = b
	case "default_view":
		if value != "source" && value != "files" && value != "history" {
			return fmt.Errorf("default_view must be source, files, or history")
		}
		cfg.DefaultView = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

// List returns all config keys and their current values.
func List() (map[string]string, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"editor":          cfg.Editor,
		"theme":           cfg.Theme,
		"poll_interval":   strconv.Itoa(cfg.PollInterval),
		"show_untracked":  strconv.FormatBool(cfg.ShowUntracked),
		"show_ignored":    strconv.FormatBool(cfg.ShowIgnored),
		"confirm_discard": strconv.FormatBool(cfg.ConfirmDiscard),
		"confirm_force":   strconv.FormatBool(cfg.ConfirmForce),
		"default_view":    cfg.DefaultView,
	}, nil
}
