package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	installDir        string
	installFileID     int
	installAcceptEULA bool
	installUseSaved   bool
	installUseArg     bool
	installNoPrompt   bool
)

var installCmd = &cobra.Command{
	Use:   "install [SOURCE]",
	Short: "Install or update a CurseForge server pack",
	Long: `Smart install/update command. If the target directory already contains
server.properties, it performs an update. Otherwise it installs fresh.

SOURCE can be a modpack ID (digits) or a CurseForge modpack URL.
If omitted, reads the saved pack ID from .mcserver/state.json.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverDir, _ := filepath.Abs(installDir)
		var source string
		if len(args) > 0 {
			source = args[0]
		}
		err := installOrUpdate(serverDir, source, installFileID, installAcceptEULA, installUseSaved, installUseArg, installNoPrompt, false)
		if err != nil {
			return handleError(cmd, err)
		}
		return nil
	},
}

func init() {
	installCmd.Flags().StringVar(&installDir, "dir", ".", "Target server directory")
	installCmd.Flags().IntVar(&installFileID, "file-id", 0, "Specific CurseForge file ID")
	installCmd.Flags().BoolVar(&installAcceptEULA, "accept-eula", false, "Write eula.txt with eula=true")
	installCmd.Flags().BoolVar(&installUseSaved, "use-saved", false, "On pack ID mismatch, prefer saved ID")
	installCmd.Flags().BoolVar(&installUseArg, "use-arg", false, "On pack ID mismatch, prefer argument ID")
	installCmd.Flags().BoolVar(&installNoPrompt, "no-prompt", false, "Fail on ambiguity instead of prompting")
}
