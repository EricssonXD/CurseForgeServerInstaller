package curseforge

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	mcerrors "github.com/ericsson/curseforge-server-installer/internal/errors"
	mchttp "github.com/ericsson/curseforge-server-installer/internal/http"
)

const baseURL = "https://api.curseforge.com"

// ModFile represents a file entry from CurseForge.
type ModFile struct {
	ID                int
	DisplayName       string
	FileDate          string
	IsServerPack      bool
	ServerPackFileID  *int
	DownloadURL       string
}

// Client interacts with the CurseForge API.
type Client struct {
	apiKey string
}

// NewClient creates a CurseForge client with the given API key.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, mcerrors.NewMissingApiKeyError("Missing CurseForge API key. Run: cfs config set-api-key")
	}
	return &Client{apiKey: apiKey}, nil
}

func (c *Client) headers() map[string]string {
	return map[string]string{
		"Accept":    "application/json",
		"x-api-key": c.apiKey,
	}
}

func (c *Client) wrapHTTPError(err error) error {
	if httpErr, ok := err.(*mchttp.HTTPError); ok {
		if httpErr.StatusCode == 403 {
			return mcerrors.NewInvalidApiKeyError(
				"CurseForge API returned 403 Forbidden (API key invalid). Update it with: cfs config set-api-key")
		}
		return mcerrors.NewUserFacingErrorf("CurseForge API request failed: HTTP %d", httpErr.StatusCode)
	}
	return err
}

// SearchModpacks searches for modpacks matching the query.
func (c *Client) SearchModpacks(query string, gameVersion string, pageSize int) ([]map[string]any, error) {
	params := map[string]string{
		"gameId":       "432",
		"classId":      "4471",
		"index":        "0",
		"pageSize":     strconv.Itoa(pageSize),
		"searchFilter": query,
		"sortField":    "2",
		"sortOrder":    "desc",
	}
	if gameVersion != "" {
		params["gameVersion"] = gameVersion
	}
	payload, err := mchttp.GetJSON(baseURL+"/v1/mods/search", c.headers(), params)
	if err != nil {
		return nil, c.wrapHTTPError(err)
	}
	data, _ := payload["data"].([]any)
	results := make([]map[string]any, 0, len(data))
	for _, item := range data {
		if m, ok := item.(map[string]any); ok {
			results = append(results, m)
		}
	}
	return results, nil
}

// ResolvePackIDFromURL extracts a pack ID from a CurseForge modpack URL.
func (c *Client) ResolvePackIDFromURL(url string) (int, error) {
	re := regexp.MustCompile(`/modpacks/([^/?#]+)`)
	match := re.FindStringSubmatch(url)
	if match == nil {
		return 0, mcerrors.NewUserFacingError("Invalid CurseForge modpack URL (expected /minecraft/modpacks/<slug>).")
	}
	slug := match[1]
	query := strings.ReplaceAll(slug, "-", " ")
	results, err := c.SearchModpacks(query, "", 5)
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, mcerrors.NewUserFacingError("No modpack found for the given URL.")
	}
	id, _ := results[0]["id"].(float64)
	return int(id), nil
}

// ListFiles returns all files for a modpack.
func (c *Client) ListFiles(packID int) ([]ModFile, error) {
	payload, err := mchttp.GetJSON(
		fmt.Sprintf("%s/v1/mods/%d/files", baseURL, packID),
		c.headers(), nil,
	)
	if err != nil {
		return nil, c.wrapHTTPError(err)
	}
	data, _ := payload["data"].([]any)
	files := make([]ModFile, 0, len(data))
	for _, item := range data {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		f := ModFile{
			ID:           intFromAny(m["id"]),
			DisplayName:  strFromAny(m["displayName"]),
			FileDate:     strFromAny(m["fileDate"]),
			IsServerPack: boolFromAny(m["isServerPack"]),
			DownloadURL:  strFromAny(m["downloadUrl"]),
		}
		if spfID := m["serverPackFileId"]; spfID != nil {
			v := intFromAny(spfID)
			if v != 0 {
				f.ServerPackFileID = &v
			}
		}
		files = append(files, f)
	}
	return files, nil
}

// GetDownloadURL returns the download URL for a specific file.
func (c *Client) GetDownloadURL(packID, fileID int) (string, error) {
	payload, err := mchttp.GetJSON(
		fmt.Sprintf("%s/v1/mods/%d/files/%d/download-url", baseURL, packID, fileID),
		c.headers(), nil,
	)
	if err != nil {
		return "", c.wrapHTTPError(err)
	}
	data, ok := payload["data"].(string)
	if !ok || data == "" {
		return "", mcerrors.NewUserFacingError("Could not resolve download URL from CurseForge API.")
	}
	return data, nil
}

// ChooseLatestServerPack returns (serverPackFileID, displayName, fileDate).
func (c *Client) ChooseLatestServerPack(packID int) (int, string, string, error) {
	files, err := c.ListFiles(packID)
	if err != nil {
		return 0, "", "", err
	}
	if len(files) == 0 {
		return 0, "", "", mcerrors.NewUserFacingError("No files found for this modpack.")
	}

	// Prefer explicit server packs
	var serverCandidates []ModFile
	for _, f := range files {
		if f.IsServerPack {
			serverCandidates = append(serverCandidates, f)
		}
	}
	if len(serverCandidates) > 0 {
		sort.Slice(serverCandidates, func(i, j int) bool {
			return serverCandidates[i].FileDate > serverCandidates[j].FileDate
		})
		f := serverCandidates[0]
		return f.ID, f.DisplayName, f.FileDate, nil
	}

	// Fallback: newest file with serverPackFileId
	sort.Slice(files, func(i, j int) bool {
		return files[i].FileDate > files[j].FileDate
	})
	newest := files[0]
	if newest.ServerPackFileID != nil {
		return *newest.ServerPackFileID, newest.DisplayName, newest.FileDate, nil
	}
	return 0, "", "", mcerrors.NewUserFacingError("No server pack found for this modpack.")
}

// ResolveServerPackDownload returns (downloadURL, serverPackFileID, displayName).
func (c *Client) ResolveServerPackDownload(packID int, fileID *int) (string, int, string, error) {
	var serverFileID int
	var displayName string

	if fileID == nil {
		var err error
		serverFileID, displayName, _, err = c.ChooseLatestServerPack(packID)
		if err != nil {
			return "", 0, "", err
		}
	} else {
		files, err := c.ListFiles(packID)
		if err != nil {
			return "", 0, "", err
		}
		var matched *ModFile
		for i := range files {
			if files[i].ID == *fileID {
				matched = &files[i]
				break
			}
		}
		if matched == nil {
			return "", 0, "", mcerrors.NewUserFacingErrorf("File id %d not found for pack %d.", *fileID, packID)
		}
		displayName = matched.DisplayName
		if matched.IsServerPack {
			serverFileID = matched.ID
		} else if matched.ServerPackFileID != nil {
			serverFileID = *matched.ServerPackFileID
		} else {
			return "", 0, "", mcerrors.NewUserFacingError("Selected file does not have an associated server pack.")
		}
	}

	url, err := c.GetDownloadURL(packID, serverFileID)
	if err != nil {
		return "", 0, "", err
	}
	return url, serverFileID, displayName, nil
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

func strFromAny(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolFromAny(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
