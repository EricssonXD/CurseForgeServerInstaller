package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	StateDirName  = ".mcserver"
	StateFileName = "state.json"
)

// stateJSON is the on-disk JSON schema.
type stateJSON struct {
	Provider             string `json:"provider"`
	PackID               *int   `json:"packId"`
	InstalledFileID      *int   `json:"installedFileId"`
	InstalledDisplayName string `json:"installedDisplayName"`
	Channel              string `json:"channel"`
	LastUpdatedAt        string `json:"lastUpdatedAt"`
}

// ServerState holds the per-server-directory installation state.
type ServerState struct {
	Provider             string
	PackID               *int
	InstalledFileID      *int
	InstalledDisplayName string
	Channel              string
	LastUpdatedAt        string
}

// UTCNowISO returns the current UTC time as an ISO 8601 string.
func UTCNowISO() string {
	return time.Now().UTC().Truncate(time.Second).Format("2006-01-02T15:04:05Z")
}

// statePath returns the full path to state.json for a given server directory.
func statePath(serverDir string) string {
	return filepath.Join(serverDir, StateDirName, StateFileName)
}

// Load reads state from <serverDir>/.mcserver/state.json.
// Returns nil, nil if the file doesn't exist.
func Load(serverDir string) (*ServerState, error) {
	path := statePath(serverDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state: %w", err)
	}
	var raw stateJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}
	s := &ServerState{
		Provider:             raw.Provider,
		PackID:               raw.PackID,
		InstalledFileID:      raw.InstalledFileID,
		InstalledDisplayName: raw.InstalledDisplayName,
		Channel:              raw.Channel,
		LastUpdatedAt:        raw.LastUpdatedAt,
	}
	if s.Provider == "" {
		s.Provider = "curseforge"
	}
	if s.Channel == "" {
		s.Channel = "latest"
	}
	return s, nil
}

// Save writes state to <serverDir>/.mcserver/state.json.
func (s *ServerState) Save(serverDir string) error {
	path := statePath(serverDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	lastUpdated := s.LastUpdatedAt
	if lastUpdated == "" {
		lastUpdated = UTCNowISO()
	}
	raw := stateJSON{
		Provider:             s.Provider,
		PackID:               s.PackID,
		InstalledFileID:      s.InstalledFileID,
		InstalledDisplayName: s.InstalledDisplayName,
		Channel:              s.Channel,
		LastUpdatedAt:        lastUpdated,
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing state: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// IntPtr is a convenience helper to create a pointer to an int.
func IntPtr(v int) *int {
	return &v
}
