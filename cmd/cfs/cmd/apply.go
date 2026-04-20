package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericsson/curseforge-server-installer/internal/download"
	"github.com/ericsson/curseforge-server-installer/internal/fs"
	"github.com/spf13/cobra"
)

var (
	applyDir       string
	applyAcceptEULA bool
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
		rawURL := args[0]

		// If it looks like a CurseForge modpack page URL, try to resolve the download URL
		if strings.Contains(rawURL, "curseforge.com/minecraft/modpacks/") {
			cf, err := getCFClient(true)
			if err != nil {
				return fmt.Errorf("CurseForge modpack URL detected but API key is needed to resolve it.\nUse 'cfs install %s' instead, or provide a direct download URL", rawURL)
			}
			packID, err := cf.ResolvePackIDFromURL(rawURL)
			if err != nil {
				return fmt.Errorf("could not resolve modpack URL: %w", err)
			}
			dlURL, _, _, err := cf.ResolveServerPackDownload(packID, nil)
			if err != nil {
				return fmt.Errorf("could not resolve download URL: %w", err)
			}
			rawURL = dlURL
			fmt.Printf("Resolved download URL from CurseForge page\n")
		}

		dir, _ := filepath.Abs(applyDir)

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
		if err := download.DownloadTo(rawURL, zipPath, "server pack"); err != nil {
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

		// Backup before replacing
		ts := strings.ReplaceAll(strings.ReplaceAll(time.Now().UTC().Format("20060102T150405Z"), ":", ""), "-", "")
		backupDir, backupErr := fs.BackupDirs(dir, ts)
		if backupErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: backup failed: %v\n", backupErr)
		} else {
			fmt.Printf("   Backup saved to %s\n", backupDir)
		}

		// Apply update
		fmt.Println("3. Replacing modpack folders...")
		if err := fs.UpdateFromPackRoot(packRoot, dir); err != nil {
			if backupDir != "" {
				fmt.Fprintln(os.Stderr, "Update failed, restoring backup...")
				if restoreErr := fs.RestoreBackup(backupDir, dir); restoreErr != nil {
					fmt.Fprintf(os.Stderr, "Restore failed: %v\n", restoreErr)
				} else {
					fmt.Fprintln(os.Stderr, "Backup restored successfully.")
				}
			}
			return fmt.Errorf("update failed: %w", err)
		}

		if applyAcceptEULA {
			os.WriteFile(filepath.Join(dir, "eula.txt"), []byte("eula=true\n"), 0o644)
		}

		fmt.Printf("--- Update complete for %s! ---\n", filepath.Base(dir))
		return nil
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyDir, "dir", "d", ".", "Server directory")
	applyCmd.Flags().BoolVar(&applyAcceptEULA, "accept-eula", false, "Write eula.txt with eula=true")
	rootCmd.AddCommand(applyCmd)
}
