package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	TLS        *TLSConfig       `json:"tls,omitempty"`
	Transport  *TransportConfig `json:"transport,omitempty"`
}

type TLSConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerName string `json:"server_name,omitempty"`
}

type TransportConfig struct {
	Type    string            `json:"type"` // ws / tcp
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type RouteConfig struct {
	Final string `json:"final"` // 默认使用的 outbound tag
}

type GenerateConfigOptions struct {
	OutPath      string
	DefaultIndex int
	InboundMode  string
	Listen       string
	SocksPort    int
	HTTPPort     int
	MixedPort    int
}

// 辅助函数，将 string 转 int，如果转换失败返回 0
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func nodeTag(prefix, name string, index int) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Sprintf("%s-%d", prefix, index)
	}

	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-", "\t", "-")
	return fmt.Sprintf("%s-%d-%s", prefix, index, replacer.Replace(name))
}

func defaultGenerateConfigOptions(outPath string, defaultIndex int) GenerateConfigOptions {
	return GenerateConfigOptions{
		OutPath:      outPath,
		DefaultIndex: defaultIndex,
		InboundMode:  "socks",
		Listen:       "127.0.0.1",
		SocksPort:    1080,
		HTTPPort:     8080,
		MixedPort:    2080,
	}
}

func DefaultGenerateConfigOptions(outPath string, defaultIndex int) GenerateConfigOptions {
	return defaultGenerateConfigOptions(outPath, defaultIndex)
}

func normalizeConfigOptions(opts GenerateConfigOptions) GenerateConfigOptions {
	defaults := defaultGenerateConfigOptions(opts.OutPath, opts.DefaultIndex)
	if opts.InboundMode == "" {
		opts.InboundMode = defaults.InboundMode
	}
	if opts.Listen == "" {
		opts.Listen = defaults.Listen
	}
	if opts.SocksPort == 0 {
		opts.SocksPort = defaults.SocksPort
	}
	if opts.HTTPPort == 0 {
		opts.HTTPPort = defaults.HTTPPort
	}
	if opts.MixedPort == 0 {
		opts.MixedPort = defaults.MixedPort
	}
	return opts
}

func buildInbounds(opts GenerateConfigOptions) ([]Inbound, error) {
	switch strings.ToLower(opts.InboundMode) {
	case "socks":
		return []Inbound{{
			Type:       "socks",
			Tag:        "socks-in",
			Listen:     opts.Listen,
			ListenPort: opts.SocksPort,
		}}, nil
	case "http":
		return []Inbound{{
			Type:       "http",
			Tag:        "http-in",
			Listen:     opts.Listen,
			ListenPort: opts.HTTPPort,
		}}, nil
	case "mixed":
		return []Inbound{{
			Type:       "mixed",
			Tag:        "mixed-in",
			Listen:     opts.Listen,
			ListenPort: opts.MixedPort,
		}}, nil
	case "both":
		return []Inbound{
			{
				Type:       "socks",
				Tag:        "socks-in",
				Listen:     opts.Listen,
				ListenPort: opts.SocksPort,
			},
			{
				Type:       "http",
				Tag:        "http-in",
				Listen:     opts.Listen,
				ListenPort: opts.HTTPPort,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported inbound mode %q, want socks, http, mixed, or both", opts.InboundMode)
	}
}

// GenerateConfig 生成 sing-box 配置文件
func GenerateConfig(vmessCfgs []VmessConfig, ssCfgs []SsConfig, outPath string, defaultIndex int) error {
	return GenerateConfigWithOptions(vmessCfgs, ssCfgs, defaultGenerateConfigOptions(outPath, defaultIndex))
}

func GenerateConfigWithOptions(vmessCfgs []VmessConfig, ssCfgs []SsConfig, opts GenerateConfigOptions) error {
	opts = normalizeConfigOptions(opts)
	outbounds := []Outbound{}

	// VMess 节点转换
	for i, v := range vmessCfgs {
		ob := Outbound{
			Type:       "vmess",
			Tag:        nodeTag("vmess", v.Ps, i+1),
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
		if strings.EqualFold(v.Tls, "tls") {
			serverName := v.Host
			if serverName == "" {
				serverName = v.Add
			}
			ob.TLS = &TLSConfig{
				Enabled:    true,
				ServerName: serverName,
			}
		}
		outbounds = append(outbounds, ob)
	}

	// SS 节点转换
	for i, s := range ssCfgs {
		ob := Outbound{
			Type:       "shadowsocks",
			Tag:        nodeTag("ss", s.Name, i+1),
			Server:     s.Server,
			ServerPort: s.Port,
			Method:     s.Method,
			Password:   s.Password,
		}
		outbounds = append(outbounds, ob)
	}

	if len(outbounds) == 0 {
		return fmt.Errorf("no supported proxy nodes found")
	}

	// 检查 defaultIndex
	if opts.DefaultIndex < 0 || opts.DefaultIndex >= len(outbounds) {
		opts.DefaultIndex = 0
	}

	inbounds, err := buildInbounds(opts)
	if err != nil {
		return err
	}

	cfg := SingboxConfig{
		Log: LogConfig{
			Level: "info",
		},
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Route: RouteConfig{
			Final: outbounds[opts.DefaultIndex].Tag,
		},
	}

	// 序列化并写入文件
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode sing-box config: %w", err)
	}

	if dir := filepath.Dir(opts.OutPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}

	if err := os.WriteFile(opts.OutPath, data, 0644); err != nil {
		return fmt.Errorf("write sing-box config: %w", err)
	}

	return nil
}
