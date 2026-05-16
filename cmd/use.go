/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var useConfigPath string

func writeSingboxConfig(path string, cfg util.SingboxConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use <id>",
	Short: "Switch sing-box route.final to a selected node.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			pterm.Error.Println("invalid node id:", args[0])
			return
		}

		cfg, err := readSingboxConfig(useConfigPath)
		if err != nil {
			pterm.Error.Println("failed to read config:", err)
			return
		}
		if len(cfg.Outbounds) == 0 {
			pterm.Warning.Println("no nodes found in config")
			return
		}
		if id < 0 || id >= len(cfg.Outbounds) {
			pterm.Error.Println(fmt.Sprintf("node id out of range: %d (available: 0-%d)", id, len(cfg.Outbounds)-1))
			return
		}

		selected := cfg.Outbounds[id]
		if selected.Tag == "" {
			pterm.Error.Println("selected node has empty tag")
			return
		}

		cfg.Route.Final = selected.Tag
		if err := writeSingboxConfig(useConfigPath, cfg); err != nil {
			pterm.Error.Println("failed to write config:", err)
			return
		}

		pterm.Success.Println(fmt.Sprintf("using node %d: %s", id, selected.Tag))
		pterm.Info.Println(fmt.Sprintf("route.final -> %s", cfg.Route.Final))
	},
}

func init() {
	rootCmd.AddCommand(useCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// useCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	useCmd.Flags().StringVarP(&useConfigPath, "config", "c", "./config.json", "sing-box config path")
}
