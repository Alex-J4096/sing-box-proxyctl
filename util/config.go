package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

// sing-box 配置
type SingboxConfig struct {
	Log       LogConfig   `json:"log"`
	Inbounds  []Inbound   `json:"inbounds"`
	Outbounds []Outbound  `json:"outbounds"`
	Route     RouteConfig `json:"route"`
}

type LogConfig struct {
	Level string `json:"level"`
}

type Inbound struct {
	Type       string `json:"type"` // "socks", "http", "tun"
	Tag        string `json:"tag,omitempty"`
	Listen     string `json:"listen"`
	ListenPort int    `json:"listen_port"`
}

type Outbound struct {
	Type       string           `json:"type"` // "vmess" 或 "shadowsocks"
	Tag        string           `json:"tag"`
	Server     string           `json:"server"`
	ServerPort int              `json:"server_port"`
	UUID       string           `json:"uuid,omitempty"`     // vmess
	Security   string           `json:"security,omitempty"` // vmess
	AlterID    int              `json:"alter_id,omitempty"` // vmess
	Method     string           `json:"method,omitempty"`   // ss
	Password   string           `json:"password,omitempty"` // ss
	Transport  *TransportConfig `json:"transport,omitempty"`
}

type TransportConfig struct {
	Type    string            `json:"type"` // ws / tcp
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type RouteConfig struct {
	Final string `json:"final"` // 默认使用的 outbound tag
}

// 辅助函数，将 string 转 int，如果转换失败返回 0
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// GenerateConfig 生成 sing-box 配置文件
func GenerateConfig(vmessCfgs []VmessConfig, ssCfgs []SsConfig, outPath string, defaultIndex int) {
	outbounds := []Outbound{}

	// VMess 节点转换
	for i, v := range vmessCfgs {
		ob := Outbound{
			Type:       "vmess",
			Tag:        fmt.Sprintf("vmess-%d", i+1),
			Server:     v.Add,
			ServerPort: mustAtoi(v.Port),
			UUID:       v.ID,
			Security:   "auto",
			AlterID:    v.Aid,
		}
		if v.Net == "ws" {
			ob.Transport = &TransportConfig{
				Type: "ws",
				Path: v.Path,
				Headers: map[string]string{
					"Host": v.Host,
				},
			}
		}
		outbounds = append(outbounds, ob)
	}

	// SS 节点转换
	for i, s := range ssCfgs {
		ob := Outbound{
			Type:       "shadowsocks",
			Tag:        fmt.Sprintf("ss-%d", i+1),
			Server:     s.Server,
			ServerPort: s.Port,
			Method:     s.Method,
			Password:   s.Password,
		}
		outbounds = append(outbounds, ob)
	}

	// 检查 defaultIndex
	if defaultIndex < 0 || defaultIndex >= len(outbounds) {
		defaultIndex = 0
	}

	cfg := SingboxConfig{
		Log: LogConfig{
			Level: "info",
		},
		Inbounds: []Inbound{
			{
				Type:       "socks",
				Tag:        "socks-in",
				Listen:     "127.0.0.1",
				ListenPort: 1080,
			},
		},
		Outbounds: outbounds,
		Route: RouteConfig{
			Final: outbounds[defaultIndex].Tag,
		},
	}

	// 序列化并写入文件
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatal("JSON encoding failed:", err)
	}

	err = os.WriteFile(outPath, data, 0644)
	if err != nil {
		log.Fatal("failed to write config.json:", err)
	}

	fmt.Printf("sing-box configuration file generated: %s\n", outPath)
}
