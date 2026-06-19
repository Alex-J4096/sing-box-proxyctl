/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var updateOutputPath string
var updateSettingsPath string
var updateInboundMode string
var updateListen string
var updateSocksPort int
var updateHTTPPort int
var updateMixedPort int
var updateSubscriptionPath string

func loadSubscriptionURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	url := strings.TrimSpace(string(data))
	if url == "" {
		return "", fmt.Errorf("subscription file is empty")
	}
	return url, nil
}

func saveSubscriptionURL(path, url string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(url)+"\n"), 0600); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

func decodeSubscription(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}

	normalized := strings.NewReplacer("\r", "", "\n", "", " ", "", "\t", "").Replace(raw)
	candidates := []string{normalized}
	if padding := len(normalized) % 4; padding != 0 {
		candidates = append(candidates, normalized+strings.Repeat("=", 4-padding))
	}

	for _, candidate := range candidates {
		if decoded, err := base64.StdEncoding.DecodeString(candidate); err == nil {
			return string(decoded)
		}
		if decoded, err := base64.URLEncoding.DecodeString(candidate); err == nil {
			return string(decoded)
		}
		if decoded, err := base64.RawStdEncoding.DecodeString(candidate); err == nil {
			return string(decoded)
		}
		if decoded, err := base64.RawURLEncoding.DecodeString(candidate); err == nil {
			return string(decoded)
		}
	}

	return raw
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull proxy subscription.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subscriptionPath := updateSubscriptionPath
		if subscriptionPath == "" {
			subscriptionPath = util.DefaultSubscriptionPath(updateOutputPath)
		}

		var subURL string
		if len(args) == 1 {
			subURL = strings.TrimSpace(args[0])
			if subURL == "" {
				return commandError("subscription URL cannot be empty")
			}
		} else {
			var err error
			subURL, err = loadSubscriptionURL(subscriptionPath)
			if err != nil {
				if os.IsNotExist(err) {
					return commandError("no saved subscription; run proxyctl sub update <url> first")
				}
				return commandError("failed to read saved subscription: %w", err)
			}
		}

		vmessCfgs := []util.VmessConfig{}
		ssCfgs := []util.SsConfig{}

		client := http.Client{Timeout: 20 * time.Second}
		resp, err := client.Get(subURL)
		if err != nil {
			return commandError("failed to fetch subscription: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return commandError("subscription request failed: HTTP %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return commandError("failed to read subscription response: %w", err)
		}
		if len(strings.TrimSpace(string(body))) == 0 {
			return commandError("subscription response is empty")
		}

		decodedStr := decodeSubscription(body)
		linkList := strings.Split(decodedStr, "\n")
		unsupportedCount := 0
		parseErrorCount := 0

		for _, line := range linkList {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 按照协议类型分类
			if strings.HasPrefix(line, "vmess://") {
				vmess, err := util.ParseVmessLink(line)
				if err != nil {
					parseErrorCount++
					pterm.Warning.Println(fmt.Sprintf("skip invalid vmess node: %v", err))
					continue
				}
				if vmess != nil {
					vmessCfgs = append(vmessCfgs, *vmess)
				}
			} else if strings.HasPrefix(line, "ss://") {
				ss, err := util.ParseSSLink(line)
				if err != nil {
					parseErrorCount++
					pterm.Warning.Println(fmt.Sprintf("skip invalid ss node: %v", err))
					continue
				}
				if ss != nil {
					ssCfgs = append(ssCfgs, *ss)
				}
			} else {
				unsupportedCount++
				pterm.Warning.Println("skip unsupported proxy type:", line)
			}
		}

		if len(vmessCfgs)+len(ssCfgs) == 0 {
			switch {
			case parseErrorCount > 0:
				return commandError("failed to parse subscription: %d invalid supported node(s), %d unsupported line(s)", parseErrorCount, unsupportedCount)
			case unsupportedCount > 0:
				return commandError("no supported proxy nodes found: %d unsupported line(s)", unsupportedCount)
			default:
				return commandError("no proxy nodes found in subscription")
			}
		}

		settingsPath := updateSettingsPath
		if settingsPath == "" {
			settingsPath = util.DefaultSettingsPath(updateOutputPath)
		}
		settings, err := util.LoadProxyctlSettings(settingsPath)
		if err != nil {
			pterm.Warning.Println("failed to read settings, using defaults:", err)
			settings = util.DefaultProxyctlSettings()
		}

		flags := cmd.Flags()
		if flags.Changed("inbound") {
			settings.Inbound.Mode = updateInboundMode
		}
		if flags.Changed("listen") {
			settings.Inbound.Listen = updateListen
		}
		if flags.Changed("socks-port") {
			settings.Inbound.SocksPort = updateSocksPort
		}
		if flags.Changed("http-port") {
			settings.Inbound.HTTPPort = updateHTTPPort
		}
		if flags.Changed("mixed-port") {
			settings.Inbound.MixedPort = updateMixedPort
		}

		configOptions := settings.GenerateConfigOptions(updateOutputPath, 0)
		if err := util.GenerateConfigWithOptions(vmessCfgs, ssCfgs, configOptions); err != nil {
			return commandError("failed to generate sing-box config: %w", err)
		}
		if err := util.SaveProxyctlSettings(settingsPath, settings); err != nil {
			pterm.Warning.Println("failed to save settings:", err)
		}
		if len(args) == 1 {
			if err := saveSubscriptionURL(subscriptionPath, subURL); err != nil {
				return commandError("config was updated but subscription URL could not be saved: %w", err)
			}
		}

		pterm.Success.Println(fmt.Sprintf("updated %d nodes and wrote sing-box config to %s", len(vmessCfgs)+len(ssCfgs), updateOutputPath))
		pterm.Info.Println("subscription: " + subscriptionPath)
		pterm.Info.Println(fmt.Sprintf("inbound=%s settings=%s", settings.Inbound.Mode, settingsPath))
		printProxyEnvHint(settings, settingsPath)
		return nil
	},
}

func init() {
	subCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	updateCmd.Flags().StringVarP(&updateOutputPath, "output", "o", util.DefaultConfigPath(), "sing-box config output path")
	updateCmd.Flags().StringVar(&updateSettingsPath, "settings", "", "proxyctl settings path")
	updateCmd.Flags().StringVar(&updateInboundMode, "inbound", "socks", "inbound mode: socks, http, mixed, both")
	updateCmd.Flags().StringVar(&updateListen, "listen", "127.0.0.1", "inbound listen address")
	updateCmd.Flags().IntVar(&updateSocksPort, "socks-port", 1080, "SOCKS inbound listen port")
	updateCmd.Flags().IntVar(&updateHTTPPort, "http-port", 8080, "HTTP inbound listen port")
	updateCmd.Flags().IntVar(&updateMixedPort, "mixed-port", 2080, "mixed inbound listen port")
	updateCmd.Flags().StringVar(&updateSubscriptionPath, "subscription-file", "", "saved subscription URL file")
}
