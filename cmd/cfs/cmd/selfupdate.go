package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ericsson/curseforge-server-installer/internal/ui"
	"github.com/spf13/cobra"
)

const githubRepo = "EricssonXD/CurseForgeServerInstaller"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update cfs to the latest release from GitHub",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Step("Checking for latest release...")

		resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo))
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("GitHub API returned %d", resp.StatusCode)
		}

		var release ghRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("failed to parse release info: %w", err)
		}

		latest := strings.TrimPrefix(release.TagName, "v")
		current := strings.TrimPrefix(Version, "v")

		if latest == current {
			ui.Success(fmt.Sprintf("Already on latest version (%s)", current))
			return nil
		}

		ui.Infof("Current: %s → Latest: %s", current, latest)

		// Find matching asset
		goos := runtime.GOOS
		goarch := runtime.GOARCH
		var assetURL string
		for _, a := range release.Assets {
			name := strings.ToLower(a.Name)
			if strings.Contains(name, goos) && strings.Contains(name, goarch) {
				assetURL = a.BrowserDownloadURL
				break
			}
		}

		if assetURL == "" {
			return fmt.Errorf("no release binary found for %s/%s — download manually from https://github.com/%s/releases", goos, goarch, githubRepo)
		}

		// Download new binary
		ui.Stepf("Downloading %s...", assetURL)
		dlResp, err := http.Get(assetURL)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		defer dlResp.Body.Close()

		if dlResp.StatusCode != 200 {
			return fmt.Errorf("download returned %d", dlResp.StatusCode)
		}

		// Write to temp file next to current binary, then rename
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("cannot determine executable path: %w", err)
		}

		tmpDir, err := os.MkdirTemp("", "cfs-update-*")
		if err != nil {
			return fmt.Errorf("cannot create temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		var binaryData io.Reader = dlResp.Body

		// Handle tar.gz archives from goreleaser
		if strings.HasSuffix(strings.ToLower(assetURL), ".tar.gz") || strings.HasSuffix(strings.ToLower(assetURL), ".tgz") {
			gr, err := gzip.NewReader(dlResp.Body)
			if err != nil {
				return fmt.Errorf("failed to decompress: %w", err)
			}
			defer gr.Close()

			tr := tar.NewReader(gr)
			found := false
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("failed to read archive: %w", err)
				}
				base := filepath.Base(hdr.Name)
				if base == "cfs" || base == "cfs.exe" {
					binaryData = tr
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("binary not found in archive")
			}
		}

		tmpPath := filepath.Join(tmpDir, "cfs")
		f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return fmt.Errorf("cannot write update: %w", err)
		}

		if _, err := io.Copy(f, binaryData); err != nil {
			f.Close()
			return fmt.Errorf("download failed: %w", err)
		}
		f.Close()

		if err := os.Rename(tmpPath, exe); err != nil {
			return fmt.Errorf("cannot replace binary: %w (try running with sudo)", err)
		}

		ui.Successf("Updated to %s!", latest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
