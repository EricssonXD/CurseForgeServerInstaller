package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericsson/curseforge-server-installer/internal/config"
	"github.com/ericsson/curseforge-server-installer/internal/curseforge"
	"github.com/ericsson/curseforge-server-installer/internal/download"
	mcerrors "github.com/ericsson/curseforge-server-installer/internal/errors"
	"github.com/ericsson/curseforge-server-installer/internal/fs"
	"github.com/ericsson/curseforge-server-installer/internal/state"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cfs",
	Short:   "CurseForge Server Installer",
	Long:    "A CLI tool for installing and updating CurseForge Minecraft server packs.",
	Version: Version,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cfCmd)
	rootCmd.AddCommand(configCmd)
}

// getCFClient loads config and creates a CurseForge client, prompting for key if needed.
func getCFClient(allowPrompt bool) (*curseforge.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cfg.CurseForgeAPIKey == "" {
		if !allowPrompt || !isInteractive() {
			return nil, mcerrors.NewMissingApiKeyError("Missing CurseForge API key. Run: cfs config set-api-key")
		}
		if err := promptAndSaveKey(); err != nil {
			return nil, err
		}
		cfg, err = config.Load()
		if err != nil {
			return nil, err
		}
	}

	client, err := curseforge.NewClient(cfg.CurseForgeAPIKey)
	if err != nil {
		return nil, err
	}

	// Validate key early
	_, validateErr := client.SearchModpacks("a", "", 1)
	if validateErr != nil {
		if _, ok := validateErr.(*mcerrors.InvalidApiKeyError); ok {
			if !allowPrompt || !isInteractive() {
				return nil, validateErr
			}
			fmt.Println("Saved API key appears invalid; please re-enter it.")
			if err := promptAndSaveKey(); err != nil {
				return nil, err
			}
			cfg, _ = config.Load()
			client, err = curseforge.NewClient(cfg.CurseForgeAPIKey)
			if err != nil {
				return nil, err
			}
			if _, err := client.SearchModpacks("a", "", 1); err != nil {
				return nil, err
			}
		} else {
			return nil, validateErr
		}
	}
	return client, nil
}

func promptAndSaveKey() error {
	fmt.Print("CurseForge API key (will be saved): ")
	reader := bufio.NewReader(os.Stdin)
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" {
		return mcerrors.NewUserFacingError("API key cannot be empty.")
	}
	cfg, _ := config.Load()
	cfg.CurseForgeAPIKey = key
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("Saved API key to %s\n", config.ConfigPath())
	return nil
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func looksLikeURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func looksLikePackID(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isServerDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "server.properties"))
	return err == nil
}

// resolvePackID figures out the pack ID from args/state, handling conflicts.
func resolvePackID(
	cf *curseforge.Client,
	source string,
	serverDir string,
	useSaved, useArg, noPrompt bool,
) (int, *state.ServerState, error) {
	saved, err := state.Load(serverDir)
	if err != nil {
		return 0, nil, err
	}
	var savedPackID *int
	if saved != nil {
		savedPackID = saved.PackID
	}

	var argPackID *int
	if source != "" {
		if looksLikeURL(source) {
			id, err := cf.ResolvePackIDFromURL(source)
			if err != nil {
				return 0, nil, err
			}
			argPackID = &id
		} else if looksLikePackID(source) {
			var id int
			fmt.Sscanf(source, "%d", &id)
			argPackID = &id
		} else {
			return 0, nil, mcerrors.NewUserFacingError("SOURCE must be a digits-only modpack id or a CurseForge modpack URL.")
		}
	}

	if savedPackID != nil && argPackID != nil && *savedPackID != *argPackID {
		if useSaved && useArg {
			return 0, nil, mcerrors.NewUserFacingError("Choose only one of --use-saved or --use-arg.")
		}
		if useSaved {
			return *savedPackID, saved, nil
		}
		if useArg {
			return *argPackID, saved, nil
		}
		if noPrompt || !isInteractive() {
			return 0, nil, mcerrors.NewUserFacingErrorf(
				"This folder is configured for packId=%d but you provided %d. Re-run with --use-saved or --use-arg (or drop --no-prompt).",
				*savedPackID, *argPackID,
			)
		}
		fmt.Printf("This folder is configured for packId=%d but you provided %d.\n", *savedPackID, *argPackID)
		fmt.Print("Use [s]aved or [a]rg? (s/a): ")
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if strings.HasPrefix(choice, "s") {
			return *savedPackID, saved, nil
		}
		if strings.HasPrefix(choice, "a") {
			return *argPackID, saved, nil
		}
		return 0, nil, mcerrors.NewUserFacingError("Invalid choice.")
	}

	if argPackID != nil {
		return *argPackID, saved, nil
	}
	if savedPackID != nil {
		return *savedPackID, saved, nil
	}
	return 0, nil, mcerrors.NewUserFacingError("No SOURCE provided and no saved packId found in .mcserver/state.json.")
}

// installOrUpdate is the core install/update logic.
func installOrUpdate(
	serverDir, source string,
	fileID int,
	acceptEULA, useSaved, useArg, noPrompt, checkOnly bool,
) error {
	cf, err := getCFClient(true)
	if err != nil {
		return err
	}

	packID, savedState, err := resolvePackID(cf, source, serverDir, useSaved, useArg, noPrompt)
	if err != nil {
		return err
	}

	modeUpdate := isServerDir(serverDir)
	mode := "install"
	if modeUpdate {
		mode = "update"
	}
	fmt.Printf("Target directory: %s\n", serverDir)
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Resolving server pack for packId=%d...\n", packID)

	var fid *int
	if fileID != 0 {
		fid = &fileID
	}
	url, serverFileID, displayName, err := cf.ResolveServerPackDownload(packID, fid)
	if err != nil {
		return err
	}

	if checkOnly && modeUpdate {
		var installed *int
		if savedState != nil {
			installed = savedState.InstalledFileID
		}
		if installed != nil && *installed == serverFileID {
			fmt.Println("Up to date.")
			return nil
		}
		fmt.Printf("Update available: installed=%v latest=%d\n", installed, serverFileID)
		return nil
	}

	// Download and extract
	tmpDir, err := os.MkdirTemp("", "mcserver_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "serverpack.zip")
	extracted := filepath.Join(tmpDir, "extracted")
	os.MkdirAll(extracted, 0o755)

	fmt.Printf("Server pack: %s (fileId=%d)\n", displayName, serverFileID)
	if err := download.DownloadTo(url, zipPath, "Downloading server pack"); err != nil {
		return err
	}
	fmt.Println("Extracting...")
	if err := fs.ExtractZip(zipPath, extracted); err != nil {
		return err
	}
	packRoot := fs.DetectPackRoot(extracted)
	fmt.Printf("Detected pack root: %s\n", packRoot)

	if modeUpdate {
		// Backup before update
		ts := state.UTCNowISO()
		ts = strings.ReplaceAll(strings.ReplaceAll(ts, ":", ""), "-", "")
		backupDir, err := fs.BackupDirs(serverDir, ts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: backup failed: %v\n", err)
		} else {
			fmt.Printf("Backup saved to %s\n", backupDir)
		}

		fmt.Println("Applying update (replacing modpack folders, preserving world/server config)...")
		if err := fs.UpdateFromPackRoot(packRoot, serverDir); err != nil {
			if backupDir != "" {
				fmt.Fprintln(os.Stderr, "Update failed, restoring backup...")
				if restoreErr := fs.RestoreBackup(backupDir, serverDir); restoreErr != nil {
					fmt.Fprintf(os.Stderr, "Restore failed: %v\n", restoreErr)
				} else {
					fmt.Fprintln(os.Stderr, "Backup restored successfully.")
				}
			}
			return err
		}
		fmt.Println("Update complete.")
	} else {
		fmt.Println("Installing into target directory...")
		if err := fs.CopyTreeContents(packRoot, serverDir); err != nil {
			return err
		}
		fmt.Println("Install complete.")
	}

	if acceptEULA {
		fmt.Println("Writing eula.txt (eula=true)...")
		os.WriteFile(filepath.Join(serverDir, "eula.txt"), []byte("eula=true\n"), 0o644)
	}

	newState := savedState
	if newState == nil {
		newState = &state.ServerState{Provider: "curseforge", Channel: "latest"}
	}
	newState.PackID = &packID
	newState.InstalledFileID = &serverFileID
	newState.InstalledDisplayName = displayName
	newState.LastUpdatedAt = state.UTCNowISO()
	if err := newState.Save(serverDir); err != nil {
		return err
	}
	fmt.Println("Saved .mcserver/state.json")
	return nil
}

// handleError prints user-facing errors to stderr and exits with code 2.
func handleError(cmd *cobra.Command, err error) error {
	if mcerrors.IsUserFacing(err) {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	return err
}
