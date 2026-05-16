/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var updateOutputPath string

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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		subURL := args[0]

		vmessCfgs := []util.VmessConfig{}
		ssCfgs := []util.SsConfig{}

		resp, err := http.Get(subURL)
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			pterm.Error.Println(fmt.Sprintf("subscription request failed: HTTP %d", resp.StatusCode))
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}

		decodedStr := decodeSubscription(body)
		linkList := strings.Split(decodedStr, "\n")

		for _, line := range linkList {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 按照协议类型分类
			if strings.HasPrefix(line, "vmess://") {
				vmess, err := util.ParseVmessLink(line)
				if err != nil {
					pterm.Warning.Println(fmt.Sprintf("skip invalid vmess node: %v", err))
					continue
				}
				if vmess != nil {
					vmessCfgs = append(vmessCfgs, *vmess)
				}
			} else if strings.HasPrefix(line, "ss://") {
				ss, err := util.ParseSSLink(line)
				if err != nil {
					pterm.Warning.Println(fmt.Sprintf("skip invalid ss node: %v", err))
					continue
				}
				if ss != nil {
					ssCfgs = append(ssCfgs, *ss)
				}
			} else {
				pterm.Warning.Println("skip unsupported proxy type:", line)
			}
		}

		// 默认使用第 0个节点
		if err := util.GenerateConfig(vmessCfgs, ssCfgs, updateOutputPath, 0); err != nil {
			pterm.Error.Println(err.Error())
			return
		}

		pterm.Success.Println(fmt.Sprintf("updated %d nodes and wrote sing-box config to %s", len(vmessCfgs)+len(ssCfgs), updateOutputPath))
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
	updateCmd.Flags().StringVarP(&updateOutputPath, "output", "o", "./config.json", "sing-box config output path")
}
