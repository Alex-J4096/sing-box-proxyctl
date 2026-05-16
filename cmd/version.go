package cmd

import (
	"fmt"
	"runtime"

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
		if !versionVerbose {
			fmt.Println("proxyctl " + Version)
			return
		}

		tableData := pterm.TableData{
			{"key", "value"},
			{"version", Version},
			{"commit", GitCommit},
			{"build_date", BuildDate},
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

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionVerbose, "verbose", "v", false, "print detailed version information")
}
