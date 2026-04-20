package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ericsson/curseforge-server-installer/internal/download"
	"github.com/ericsson/curseforge-server-installer/internal/fs"
	"github.com/spf13/cobra"
)

var (
	applyDir string
)

var applyCmd = &cobra.Command{
	Use:   "apply <URL>",
	Short: "Apply a server pack ZIP from a direct download URL (no API key needed)",
	Long: `Downloads a server pack ZIP from a direct URL and applies it to the server directory.
This is a fallback for users without a CurseForge API key. It replaces modpack-managed
folders (mods, config, scripts, kubejs, libraries, defaultconfigs) while preserving
world data, server.properties, and other server-specific files.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		dir := applyDir
		if dir == "" {
			dir = "."
		}
		dir, _ = filepath.Abs(dir)

		// Safety check
		if _, err := os.Stat(filepath.Join(dir, "server.properties")); err != nil {
			return fmt.Errorf("'server.properties' not found in %s — are you in the right folder?", dir)
		}

		// Download to temp file
		tmpDir, err := os.MkdirTemp("", "cfs-apply-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		zipPath := filepath.Join(tmpDir, "update.zip")
		fmt.Println("1. Downloading server pack...")
		if err := download.DownloadTo(url, zipPath, "server pack"); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		// Extract
		extractDir := filepath.Join(tmpDir, "extracted")
		fmt.Println("2. Extracting files...")
		if err := fs.ExtractZip(zipPath, extractDir); err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}

		// Detect pack root
		packRoot := fs.DetectPackRoot(extractDir)

		// Apply update
		fmt.Println("3. Replacing modpack folders...")
		if err := fs.UpdateFromPackRoot(packRoot, dir); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Printf("--- Update complete for %s! ---\n", filepath.Base(dir))
		return nil
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyDir, "dir", "d", "", "Server directory (default: current directory)")
	rootCmd.AddCommand(applyCmd)
}
