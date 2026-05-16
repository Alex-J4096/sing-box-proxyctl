/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var listConfigPath string
var listPingCachePath string

func outboundTransport(outbound util.Outbound) string {
	if outbound.Transport == nil || outbound.Transport.Type == "" {
		return "tcp"
	}
	return outbound.Transport.Type
}

func outboundTLS(outbound util.Outbound) string {
	if outbound.TLS != nil && outbound.TLS.Enabled {
		return "on"
	}
	return "off"
}

func displayTag(tag string) string {
	if tag == "" {
		return "-"
	}
	return tag
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List nodes from sing-box config.",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(listConfigPath)
		if err != nil {
			pterm.Error.Println("failed to read config:", err)
			return
		}

		var cfg util.SingboxConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			pterm.Error.Println("failed to parse config:", err)
			return
		}

		if len(cfg.Outbounds) == 0 {
			pterm.Warning.Println("no nodes found in config")
			return
		}

		titleStyle := pterm.NewStyle(pterm.FgLightGreen, pterm.Bold)
		activeStyle := pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)
		mutedStyle := pterm.NewStyle(pterm.FgGray)

		titleStyle.Println("proxyctl::node-list")
		mutedStyle.Println("config: " + listConfigPath)

		pingCachePath := listPingCachePath
		if pingCachePath == "" {
			pingCachePath = util.DefaultPingCachePath(listConfigPath)
		}
		pingCache, err := util.LoadPingCache(pingCachePath)
		if err != nil {
			pterm.Warning.Println("failed to read ping cache:", err)
		} else {
			mutedStyle.Println("ping-cache: " + pingCachePath)
		}

		tableData := pterm.TableData{
			{"id", "use", "type", "tag", "region", "latency", "status", "transport", "tls"},
		}

		for i, outbound := range cfg.Outbounds {
			current := ""
			if outbound.Tag == cfg.Route.Final {
				current = activeStyle.Sprint("*")
			}

			region := "-"
			latency := "-"
			status := "-"
			if result, ok := pingCache.Results[util.PingCacheKey(outbound, i)]; ok && util.MatchPingResult(outbound, result) {
				region = util.FormatRegion(result.Region)
				latency = util.FormatLatency(result)
				status = result.Status
			}

			tableData = append(tableData, []string{
				strconv.Itoa(i),
				current,
				outbound.Type,
				displayTag(outbound.Tag),
				region,
				latency,
				status,
				outboundTransport(outbound),
				outboundTLS(outbound),
			})
		}

		if err := pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(tableData).
			Render(); err != nil {
			pterm.Error.Println("failed to render node list:", err)
			return
		}
	},
}

func init() {
	nodeCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listCmd.Flags().StringVarP(&listConfigPath, "config", "c", "./config.json", "sing-box config path")
	listCmd.Flags().StringVar(&listPingCachePath, "ping-cache", "", "ping result cache path")
}
