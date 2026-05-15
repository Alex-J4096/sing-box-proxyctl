/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull proxy subscription.",
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]

		vmessCfgs := []util.VmessConfig{}
		ssCfgs := []util.SsConfig{}

		resp, err := http.Get(url)
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}

		// 处理空白字符并解码
		b64Str := strings.TrimSpace(string(body))
		decoded, _ := base64.StdEncoding.DecodeString(string(b64Str))
		if err != nil {
			pterm.Error.Println(err.Error())
			return
		}

		// 按 utf-8解码并按行分割
		decodedStr := string(decoded)
		linkList := strings.Split(decodedStr, "\n")

		for _, line := range linkList {
			//fmt.Println(line)
			// 按照协议类型分类
			if strings.HasPrefix(line, "vmess://") {
				vmess, _ := util.ParseVmessLink(line)
				if vmess != nil {
					vmessCfgs = append(vmessCfgs, *vmess)
				}
			} else if strings.HasPrefix(line, "ss://") {
				ss, _ := util.ParseSSLink(line)
				if ss != nil {
					ssCfgs = append(ssCfgs, *ss)
				}
			} else {
				pterm.Error.Print("Unexpected proxy types:", line)
			}
		}
		// 写入config文件
		outPath := "./config.json"
		// 默认使用第 0个节点
		util.GenerateConfig(vmessCfgs, ssCfgs, outPath, 0)
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
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
