package cmd

import (
	"fmt"

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
		cf, err := getCFClient(true)
		if err != nil {
			return handleError(cmd, err)
		}
		packID, err := cf.ResolvePackIDFromURL(args[0])
		if err != nil {
			return handleError(cmd, err)
		}
		fmt.Println(packID)
		return nil
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
		cf, err := getCFClient(true)
		if err != nil {
			return handleError(cmd, err)
		}
		results, err := cf.SearchModpacks(args[0], cfSearchGameVersion, cfSearchLimit)
		if err != nil {
			return handleError(cmd, err)
		}
		for _, item := range results {
			id := item["id"]
			name := item["name"]
			fmt.Printf("%d\t%v\n", int(id.(float64)), name)
		}
		return nil
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
		cf, err := getCFClient(true)
		if err != nil {
			return handleError(cmd, err)
		}
		var packID int
		fmt.Sscanf(args[0], "%d", &packID)
		files, err := cf.ListFiles(packID)
		if err != nil {
			return handleError(cmd, err)
		}
		count := 0
		for _, f := range files {
			if count >= cfFilesLimit {
				break
			}
			if cfFilesServerOnly && !f.IsServerPack && f.ServerPackFileID == nil {
				continue
			}
			spfID := 0
			if f.ServerPackFileID != nil {
				spfID = *f.ServerPackFileID
			}
			fmt.Printf("%d\t%s\t%s\tserverPack=%v\tserverPackFileId=%d\n",
				f.ID, f.FileDate, f.DisplayName, f.IsServerPack, spfID)
			count++
		}
		return nil
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
		cf, err := getCFClient(true)
		if err != nil {
			return handleError(cmd, err)
		}
		var packID int
		fmt.Sscanf(args[0], "%d", &packID)
		var fid *int
		if cfDlFileID != 0 {
			fid = &cfDlFileID
		}
		url, serverFileID, displayName, err := cf.ResolveServerPackDownload(packID, fid)
		if err != nil {
			return handleError(cmd, err)
		}
		if cfDlVerbose {
			fmt.Printf("serverPackFileId=%d\tdisplayName=%s\n", serverFileID, displayName)
		}
		fmt.Println(url)
		return nil
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
