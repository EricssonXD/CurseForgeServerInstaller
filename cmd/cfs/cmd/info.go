package cmd

import (
	"fmt"
	"strings"

	"github.com/ericsson/curseforge-server-installer/internal/ui"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <PACK_ID or URL>",
	Short: "Show details about a CurseForge modpack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cf, err := getCFClient(true)
		if err != nil {
			return handleError(cmd, err)
		}

		source := args[0]
		packID, _, err := resolvePackID(cf, source, "", false, false, true)
		if err != nil {
			return handleError(cmd, err)
		}

		mod, err := cf.GetModpack(packID)
		if err != nil {
			return handleError(cmd, err)
		}

		name, _ := mod["name"].(string)
		summary, _ := mod["summary"].(string)
		dlCount, _ := mod["downloadCount"].(float64)

		ui.Stepf("📦 %s", name)
		ui.Infof("   ID: %d", packID)
		if summary != "" {
			ui.Infof("   %s", summary)
		}
		ui.Infof("   Downloads: %s", formatCount(dlCount))

		// Game versions
		if versions, ok := mod["latestFilesIndexes"].([]any); ok && len(versions) > 0 {
			seen := map[string]bool{}
			var gameVersions []string
			for _, v := range versions {
				if m, ok := v.(map[string]any); ok {
					if gv, ok := m["gameVersion"].(string); ok && gv != "" && !seen[gv] {
						seen[gv] = true
						gameVersions = append(gameVersions, gv)
					}
				}
			}
			if len(gameVersions) > 0 {
				ui.Infof("   Game versions: %s", strings.Join(gameVersions, ", "))
			}
		}

		// Links
		if links, ok := mod["links"].(map[string]any); ok {
			if url, ok := links["websiteUrl"].(string); ok && url != "" {
				ui.Dimf("   %s", url)
			}
		}

		return nil
	},
}

func formatCount(n float64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", n/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", n/1_000)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
