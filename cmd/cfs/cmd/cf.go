package cmd

import (
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:   "cf",
	Short: "CurseForge helper commands",
	Long:  "Subcommands for interacting with the CurseForge API directly.",
}

// -- cf resolve ---------------------------------------------------------------

var cfResolveCmd = &cobra.Command{
	Use:   "resolve <URL>",
	Short: "Resolve a CurseForge modpack URL to a pack ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPython(buildCfPythonArgs("resolve", args, nil, nil))
	},
}

// -- cf search ----------------------------------------------------------------

var (
	cfSearchGameVersion string
	cfSearchLimit       int
)

var cfSearchCmd = &cobra.Command{
	Use:   "search <QUERY>",
	Short: "Search CurseForge modpacks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := map[string]string{
			"game-version": cfSearchGameVersion,
			"limit":        intToStr(cfSearchLimit),
		}
		return runPython(buildCfPythonArgs("search", args, flags, nil))
	},
}

// -- cf files -----------------------------------------------------------------

var (
	cfFilesServerOnly bool
	cfFilesLimit      int
)

var cfFilesCmd = &cobra.Command{
	Use:   "files <PACK_ID>",
	Short: "List files for a CurseForge modpack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := map[string]string{
			"limit": intToStr(cfFilesLimit),
		}
		boolFlags := map[string]bool{
			"server-only": cfFilesServerOnly,
		}
		return runPython(buildCfPythonArgs("files", args, flags, boolFlags))
	},
}

// -- cf download-url ----------------------------------------------------------

var (
	cfDlFileID  int
	cfDlVerbose bool
)

var cfDlCmd = &cobra.Command{
	Use:   "download-url <PACK_ID>",
	Short: "Resolve direct download URL for a server pack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := map[string]string{
			"file-id": intToStr(cfDlFileID),
		}
		boolFlags := map[string]bool{
			"verbose": cfDlVerbose,
		}
		return runPython(buildCfPythonArgs("download-url", args, flags, boolFlags))
	},
}

func init() {
	cfCmd.AddCommand(cfResolveCmd)

	cfSearchCmd.Flags().StringVar(&cfSearchGameVersion, "game-version", "", "Filter by Minecraft version")
	cfSearchCmd.Flags().IntVar(&cfSearchLimit, "limit", 10, "Max results")
	cfCmd.AddCommand(cfSearchCmd)

	cfFilesCmd.Flags().BoolVar(&cfFilesServerOnly, "server-only", false, "Only show server packs")
	cfFilesCmd.Flags().IntVar(&cfFilesLimit, "limit", 20, "Max results")
	cfCmd.AddCommand(cfFilesCmd)

	cfDlCmd.Flags().IntVar(&cfDlFileID, "file-id", 0, "Specific file to resolve")
	cfDlCmd.Flags().BoolVar(&cfDlVerbose, "verbose", false, "Print extra metadata")
	cfCmd.AddCommand(cfDlCmd)
}
