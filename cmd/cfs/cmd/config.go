package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage mcserver configuration",
	Long:  "Subcommands for managing the local mcserver config (API keys, etc.).",
}

// -- config set-api-key -------------------------------------------------------

var configSetAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key [API_KEY]",
	Short: "Save CurseForge API key to config file",
	Long:  "If API_KEY is omitted, you will be prompted interactively.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPython(buildConfigPythonArgs("set-api-key", args))
	},
}

// -- config show --------------------------------------------------------------

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current config (secrets masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPython(buildConfigPythonArgs("show", nil))
	},
}

// -- config unset-api-key -----------------------------------------------------

var configUnsetAPIKeyCmd = &cobra.Command{
	Use:   "unset-api-key",
	Short: "Remove saved CurseForge API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPython(buildConfigPythonArgs("unset-api-key", nil))
	},
}

// -- config path --------------------------------------------------------------

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPython(buildConfigPythonArgs("path", nil))
	},
}

func init() {
	configCmd.AddCommand(configSetAPIKeyCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configUnsetAPIKeyCmd)
	configCmd.AddCommand(configPathCmd)
}
