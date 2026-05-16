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

var (
	useConfigPath string
	useNoRestart  bool
	useCorePath   string
	usePidFile    string
	useLogFile    string
)

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
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return commandError("invalid node id: %s", args[0])
		}

		cfg, err := readSingboxConfig(useConfigPath)
		if err != nil {
			return commandError("failed to read config: %w", err)
		}
		if len(cfg.Outbounds) == 0 {
			return commandError("no nodes found in config")
		}
		if id < 0 || id >= len(cfg.Outbounds) {
			return commandError("node id out of range: %d (available: 0-%d)", id, len(cfg.Outbounds)-1)
		}

		selected := cfg.Outbounds[id]
		if selected.Tag == "" {
			return commandError("selected node has empty tag")
		}

		cfg.Route.Final = selected.Tag
		if err := writeSingboxConfig(useConfigPath, cfg); err != nil {
			return commandError("failed to write config: %w", err)
		}

		pterm.Success.Println(fmt.Sprintf("using node %d: %s", id, selected.Tag))
		pterm.Info.Println(fmt.Sprintf("route.final -> %s", cfg.Route.Final))

		if useNoRestart {
			pterm.Info.Println("skip sing-box restart")
			return nil
		}

		if err := restartSingbox(useConfigPath, useCorePath, usePidFile, useLogFile); err != nil {
			return commandError("failed to restart sing-box: %w", err)
		}
		return nil
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
	useCmd.Flags().BoolVar(&useNoRestart, "no-restart", false, "only update config without restarting sing-box")
	useCmd.Flags().StringVar(&useCorePath, "core", "", "sing-box core path used for restart")
	useCmd.Flags().StringVar(&usePidFile, "pid-file", "", "managed sing-box pid file used for restart")
	useCmd.Flags().StringVar(&useLogFile, "log-file", "", "managed sing-box log file used for restart")
}
