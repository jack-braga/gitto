package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Editor != "" {
		t.Errorf("expected empty editor, got %q", cfg.Editor)
	}
	if cfg.Theme != "auto" {
		t.Errorf("expected theme=auto, got %q", cfg.Theme)
	}
	if cfg.PollInterval != 5 {
		t.Errorf("expected poll_interval=5, got %d", cfg.PollInterval)
	}
	if !cfg.ShowUntracked {
		t.Error("expected show_untracked=true")
	}
	if cfg.ShowIgnored {
		t.Error("expected show_ignored=false")
	}
	if !cfg.ConfirmDiscard {
		t.Error("expected confirm_discard=true")
	}
	if !cfg.ConfirmForce {
		t.Error("expected confirm_force=true")
	}
	if cfg.DefaultView != "source" {
		t.Errorf("expected default_view=source, got %q", cfg.DefaultView)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp directory as XDG_CONFIG_HOME
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := DefaultConfig()
	cfg.Editor = "nvim"
	cfg.PollInterval = 10

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmp, "gitto", "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Editor != "nvim" {
		t.Errorf("expected editor=nvim, got %q", loaded.Editor)
	}
	if loaded.PollInterval != 10 {
		t.Errorf("expected poll_interval=10, got %d", loaded.PollInterval)
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}
	if cfg.Theme != "auto" {
		t.Errorf("expected defaults, got theme=%q", cfg.Theme)
	}
}

func TestGetAndSet(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Set a value
	if err := Set("editor", "code"); err != nil {
		t.Fatalf("Set editor failed: %v", err)
	}

	val, err := Get("editor")
	if err != nil {
		t.Fatalf("Get editor failed: %v", err)
	}
	if val != "code" {
		t.Errorf("expected code, got %q", val)
	}

	// Set poll_interval (int conversion)
	if err := Set("poll_interval", "15"); err != nil {
		t.Fatalf("Set poll_interval failed: %v", err)
	}
	val, _ = Get("poll_interval")
	if val != "15" {
		t.Errorf("expected 15, got %q", val)
	}

	// Set bool
	if err := Set("show_untracked", "false"); err != nil {
		t.Fatalf("Set show_untracked failed: %v", err)
	}
	val, _ = Get("show_untracked")
	if val != "false" {
		t.Errorf("expected false, got %q", val)
	}
}

func TestSetValidation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	tests := []struct {
		key, value string
		wantErr    bool
	}{
		{"theme", "dark", false},
		{"theme", "invalid", true},
		{"poll_interval", "0", true},
		{"poll_interval", "abc", true},
		{"show_untracked", "yes", true},
		{"default_view", "source", false},
		{"default_view", "invalid", true},
		{"unknown_key", "value", true},
	}

	for _, tt := range tests {
		err := Set(tt.key, tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("Set(%q, %q): err=%v, wantErr=%v", tt.key, tt.value, err, tt.wantErr)
		}
	}
}
