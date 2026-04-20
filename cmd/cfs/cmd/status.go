package cmd

import (
	"fmt"
	"path/filepath"

	mcerrors "github.com/ericsson/curseforge-server-installer/internal/errors"
	"github.com/ericsson/curseforge-server-installer/internal/state"
	"github.com/spf13/cobra"
)

var statusDir string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed pack/version info for a server directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverDir, _ := filepath.Abs(statusDir)
		s, err := state.Load(serverDir)
		if err != nil {
			return err
		}
		if s == nil || s.PackID == nil {
			return handleError(cmd, mcerrors.NewUserFacingError("No .mcserver/state.json found in this folder."))
		}
		fmt.Printf("packId=%d\n", *s.PackID)
		if s.InstalledFileID != nil {
			fmt.Printf("installedFileId=%d\n", *s.InstalledFileID)
		}
		fmt.Printf("installedDisplayName=%s\n", s.InstalledDisplayName)
		fmt.Printf("lastUpdatedAt=%s\n", s.LastUpdatedAt)
		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusDir, "dir", ".", "Server directory to inspect")
}
