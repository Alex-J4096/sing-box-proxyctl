package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var envConfigPath string
var envSettingsPath string

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print shell proxy environment commands for the current inbound.",
	RunE: func(cmd *cobra.Command, args []string) error {
		settingsPath := envSettingsPath
		if settingsPath == "" {
			settingsPath = util.DefaultSettingsPath(envConfigPath)
		}

		settings, err := util.LoadProxyctlSettings(settingsPath)
		if err != nil {
			return commandError("failed to read settings: %w", err)
		}

		printProxyEnvHint(settings, settingsPath)
		return nil
	},
}

func printProxyEnvHint(settings util.ProxyctlSettings, settingsPath string) {
	settings = settings.Normalize()

	pterm.Info.Println(fmt.Sprintf("inbound=%s listen=%s settings=%s", settings.Inbound.Mode, settings.Inbound.Listen, settingsPath))
	pterm.Println("Run these commands in your shell:")
	for _, line := range proxyEnvLines(settings) {
		pterm.Println(line)
	}
	pterm.Println("")
	pterm.Println("To persist them, add the lines above to ~/.zshrc or ~/.bashrc.")
}

func proxyEnvLines(settings util.ProxyctlSettings) []string {
	inbound := settings.Normalize().Inbound
	listen := inbound.Listen
	if listen == "" {
		listen = "127.0.0.1"
	}

	switch strings.ToLower(inbound.Mode) {
	case "http":
		httpAddr := httpProxyURL(listen, inbound.HTTPPort)
		return []string{
			"export HTTP_PROXY=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr,
		}
	case "mixed":
		httpAddr := httpProxyURL(listen, inbound.MixedPort)
		socksAddr := socksProxyURL(listen, inbound.MixedPort)
		return []string{
			"export ALL_PROXY=" + socksAddr,
			"export HTTP_PROXY=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr,
		}
	case "both":
		socksAddr := socksProxyURL(listen, inbound.SocksPort)
		httpAddr := httpProxyURL(listen, inbound.HTTPPort)
		return []string{
			"export ALL_PROXY=" + socksAddr,
			"export HTTP_PROXY=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr,
		}
	default:
		return []string{
			"export ALL_PROXY=" + socksProxyURL(listen, inbound.SocksPort),
		}
	}
}

func socksProxyURL(host string, port int) string {
	return "socks5h://" + host + ":" + strconv.Itoa(port)
}

func httpProxyURL(host string, port int) string {
	return "http://" + host + ":" + strconv.Itoa(port)
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.Flags().StringVarP(&envConfigPath, "config", "c", util.DefaultConfigPath(), "sing-box config path")
	envCmd.Flags().StringVar(&envSettingsPath, "settings", "", "proxyctl settings path")
}
