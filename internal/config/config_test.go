package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cfg := &AppConfig{CurseForgeAPIKey: "test-key-12345"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.CurseForgeAPIKey != "test-key-12345" {
		t.Errorf("got key %q, want %q", loaded.CurseForgeAPIKey, "test-key-12345")
	}

	// Verify file permissions
	info, _ := os.Stat(filepath.Join(tmp, "mcserver", "config.json"))
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file perm = %o, want 600", perm)
	}
}

func TestLoadMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.CurseForgeAPIKey != "" {
		t.Errorf("expected empty key, got %q", cfg.CurseForgeAPIKey)
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "(not set)"},
		{"abc", "***"},
		{"abcdefgh", "ab***gh"},
	}
	for _, tt := range tests {
		if got := MaskSecret(tt.in); got != tt.want {
			t.Errorf("MaskSecret(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := ConfigPath()
	want := "/custom/config/mcserver/config.json"
	if got != want {
		t.Errorf("ConfigPath() = %q, want %q", got, want)
	}
}
