package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionVerbose bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print proxyctl version information.",
	Run: func(cmd *cobra.Command, args []string) {
		version := buildValueOrDefault(Version, "dev")
		commit := buildValueOrDefault(GitCommit, "unknown")
		buildDate := buildValueOrDefault(BuildDate, "unknown")

		if !versionVerbose {
			fmt.Println("proxyctl " + version)
			return
		}

		tableData := pterm.TableData{
			{"key", "value"},
			{"version", version},
			{"commit", commit},
			{"build_date", buildDate},
			{"go", runtime.Version()},
			{"platform", runtime.GOOS + "/" + runtime.GOARCH},
		}

		if err := pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(tableData).
			Render(); err != nil {
			pterm.Error.Println("failed to render version:", err)
		}
	},
}

func buildValueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionVerbose, "verbose", "v", false, "print detailed version information")
}
