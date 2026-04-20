package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	updateDir        string
	updateFileID     int
	updateAcceptEULA bool
	updateCheckOnly  bool
	updateUseSaved   bool
	updateUseArg     bool
	updateNoPrompt   bool
)

var updateCmd = &cobra.Command{
	Use:   "update [SOURCE]",
	Short: "Update a CurseForge server pack (alias of install)",
	Long: `Update command — functionally identical to install, with the addition
of --check-only to see if an update is available without applying it.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverDir, _ := filepath.Abs(updateDir)
		var source string
		if len(args) > 0 {
			source = args[0]
		}
		err := installOrUpdate(serverDir, source, updateFileID, updateAcceptEULA, updateUseSaved, updateUseArg, updateNoPrompt, updateCheckOnly)
		if err != nil {
			return handleError(cmd, err)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateDir, "dir", ".", "Target server directory")
	updateCmd.Flags().IntVar(&updateFileID, "file-id", 0, "Specific CurseForge file ID")
	updateCmd.Flags().BoolVar(&updateAcceptEULA, "accept-eula", false, "Write eula.txt with eula=true")
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check-only", false, "Only check if update is available")
	updateCmd.Flags().BoolVar(&updateUseSaved, "use-saved", false, "On pack ID mismatch, prefer saved ID")
	updateCmd.Flags().BoolVar(&updateUseArg, "use-arg", false, "On pack ID mismatch, prefer argument ID")
	updateCmd.Flags().BoolVar(&updateNoPrompt, "no-prompt", false, "Fail on ambiguity instead of prompting")
}
