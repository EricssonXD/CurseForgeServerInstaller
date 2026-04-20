package cmd

import (
	"github.com/spf13/cobra"
)

var statusDir string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed pack/version info for a server directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := map[string]string{
			"dir": statusDir,
		}
		return runPython(buildPythonArgs("status", nil, flags, nil))
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusDir, "dir", ".", "Server directory to inspect")
}
