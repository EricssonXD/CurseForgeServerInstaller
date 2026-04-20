package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ericsson/curseforge-server-installer/internal/config"
	mcerrors "github.com/ericsson/curseforge-server-installer/internal/errors"
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
		var apiKey string
		if len(args) > 0 {
			apiKey = args[0]
		}
		if apiKey == "" {
			if !isInteractive() {
				return handleError(cmd, mcerrors.NewUserFacingError("No API key provided. Pass it as an argument or run interactively."))
			}
			fmt.Print("CurseForge API key: ")
			reader := bufio.NewReader(os.Stdin)
			key, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(key)
		}
		if apiKey == "" {
			return handleError(cmd, mcerrors.NewUserFacingError("API key cannot be empty."))
		}
		cfg, _ := config.Load()
		cfg.CurseForgeAPIKey = apiKey
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Saved CurseForge API key to %s\n", config.ConfigPath())
		return nil
	},
}

// -- config show --------------------------------------------------------------

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current config (secrets masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Printf("configPath=%s\n", config.ConfigPath())
		fmt.Printf("curseforgeApiKey=%s\n", config.MaskSecret(cfg.CurseForgeAPIKey))
		return nil
	},
}

// -- config unset-api-key -----------------------------------------------------

var configUnsetAPIKeyCmd = &cobra.Command{
	Use:   "unset-api-key",
	Short: "Remove saved CurseForge API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		cfg.CurseForgeAPIKey = ""
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Cleared CurseForge API key in %s\n", config.ConfigPath())
		return nil
	},
}

// -- config path --------------------------------------------------------------

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(config.ConfigPath())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetAPIKeyCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configUnsetAPIKeyCmd)
	configCmd.AddCommand(configPathCmd)
}
