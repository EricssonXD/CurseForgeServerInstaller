package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	s := &ServerState{
		Provider:             "curseforge",
		PackID:               IntPtr(12345),
		InstalledFileID:      IntPtr(67890),
		InstalledDisplayName: "Test Pack",
		Channel:              "latest",
		LastUpdatedAt:        "2024-01-01T00:00:00Z",
	}
	if err := s.Save(tmp); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *loaded.PackID != 12345 {
		t.Errorf("PackID = %d, want 12345", *loaded.PackID)
	}
	if *loaded.InstalledFileID != 67890 {
		t.Errorf("InstalledFileID = %d, want 67890", *loaded.InstalledFileID)
	}
	if loaded.InstalledDisplayName != "Test Pack" {
		t.Errorf("DisplayName = %q", loaded.InstalledDisplayName)
	}
}

func TestLoadMissing(t *testing.T) {
	s, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s != nil {
		t.Errorf("expected nil state for missing file")
	}
}

func TestLoadDefaults(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, StateDirName)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, StateFileName), []byte(`{"packId": 1}`), 0o644)

	s, err := Load(tmp)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Provider != "curseforge" {
		t.Errorf("Provider = %q, want curseforge", s.Provider)
	}
	if s.Channel != "latest" {
		t.Errorf("Channel = %q, want latest", s.Channel)
	}
}

func TestIntPtr(t *testing.T) {
	p := IntPtr(42)
	if *p != 42 {
		t.Errorf("IntPtr(42) = %d", *p)
	}
}

func TestUTCNowISO(t *testing.T) {
	s := UTCNowISO()
	if len(s) != 20 || s[len(s)-1] != 'Z' {
		t.Errorf("UTCNowISO() = %q, unexpected format", s)
	}
}
