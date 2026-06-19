package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var envConfigPath string
var envSettingsPath string
var envPidFile string
var envShellInit bool

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
			// env output is evaluated by the parent shell. Clear stale proxy
			// variables even when settings are unreadable, and keep diagnostics
			// out of the generated shell program.
			fmt.Fprintln(cmd.OutOrStdout(), proxyUnsetEnvLines()[0])
			return fmt.Errorf("failed to read settings: %w", err)
		}

		if envShellInit {
			printShellInit(envConfigPath, settingsPath, envPidFile)
			return nil
		}

		status := inspectManagedSingbox(envConfigPath, envPidFile, "")
		lines := proxyUnsetEnvLines()
		if status.Running {
			lines = proxyEnvLines(settings)
		}
		for _, line := range lines {
			fmt.Fprintln(cmd.OutOrStdout(), line)
		}
		return nil
	},
}

func printProxyEnvHint(settings util.ProxyctlSettings, settingsPath string) {
	pterm.Info.Println(fmt.Sprintf("inbound=%s listen=%s settings=%s", settings.Inbound.Mode, settings.Inbound.Listen, settingsPath))
	pterm.Info.Println(`For automatic synchronization, add this to ~/.zshrc or ~/.bashrc: eval "$(proxyctl env --shell-init)"`)
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
			proxyUnsetEnvLines()[0],
			"export HTTP_PROXY=" + httpAddr + " http_proxy=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr + " https_proxy=" + httpAddr,
		}
	case "mixed":
		httpAddr := httpProxyURL(listen, inbound.MixedPort)
		socksAddr := socksProxyURL(listen, inbound.MixedPort)
		return []string{
			proxyUnsetEnvLines()[0],
			"export ALL_PROXY=" + socksAddr + " all_proxy=" + socksAddr,
			"export HTTP_PROXY=" + httpAddr + " http_proxy=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr + " https_proxy=" + httpAddr,
		}
	case "both":
		socksAddr := socksProxyURL(listen, inbound.SocksPort)
		httpAddr := httpProxyURL(listen, inbound.HTTPPort)
		return []string{
			proxyUnsetEnvLines()[0],
			"export ALL_PROXY=" + socksAddr + " all_proxy=" + socksAddr,
			"export HTTP_PROXY=" + httpAddr + " http_proxy=" + httpAddr,
			"export HTTPS_PROXY=" + httpAddr + " https_proxy=" + httpAddr,
		}
	default:
		return []string{
			proxyUnsetEnvLines()[0],
			"export ALL_PROXY=" + socksProxyURL(listen, inbound.SocksPort) + " all_proxy=" + socksProxyURL(listen, inbound.SocksPort),
		}
	}
}

func proxyUnsetEnvLines() []string {
	return []string{"unset ALL_PROXY HTTP_PROXY HTTPS_PROXY all_proxy http_proxy https_proxy"}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

func printShellInit(configPath, settingsPath, pidFile string) {
	args := "-c " + shellQuote(configPath) + " --settings " + shellQuote(settingsPath)
	if pidFile != "" {
		args += " --pid-file " + shellQuote(pidFile)
	}
	fmt.Printf("_proxyctl_sync_env() { eval \"$(command proxyctl env %s)\"; }\n", args)
	shell := filepath.Base(os.Getenv("SHELL"))
	if shell == "zsh" {
		fmt.Println("typeset -ga precmd_functions")
		fmt.Println("(( ${precmd_functions[(I)_proxyctl_sync_env]} )) || precmd_functions+=(_proxyctl_sync_env)")
	} else {
		fmt.Println("case \";${PROMPT_COMMAND:-};\" in *\";_proxyctl_sync_env;\"*) ;; *) PROMPT_COMMAND=\"_proxyctl_sync_env${PROMPT_COMMAND:+;$PROMPT_COMMAND}\" ;; esac")
	}
	fmt.Println("_proxyctl_sync_env")
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
	envCmd.Flags().StringVar(&envPidFile, "pid-file", "", "managed sing-box pid file")
	envCmd.Flags().BoolVar(&envShellInit, "shell-init", false, "print a zsh/bash hook that keeps proxy variables in sync")
}
