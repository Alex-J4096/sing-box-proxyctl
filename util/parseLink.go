package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type VmessConfig struct {
	V    string `json:"v"`              // 版本号
	Ps   string `json:"ps"`             // 节点名称
	Add  string `json:"add"`            // 服务器地址
	Port string `json:"port"`           // 端口
	ID   string `json:"id"`             // UUID
	Aid  int    `json:"aid"`            // alterId
	Net  string `json:"net"`            // 网络类型 ws/tcp
	Type string `json:"type"`           // 伪装类型
	Host string `json:"host,omitempty"` // ws host
	Path string `json:"path,omitempty"` // ws path
	Tls  string `json:"tls"`            // tls/tcp
}

type SsConfig struct {
	Name       string `json:"name"`        // 节点名称
	Server     string `json:"server"`      // 服务器地址
	Port       int    `json:"port"`        // 端口
	Method     string `json:"method"`      // 加密方式
	Password   string `json:"password"`    // 密码
	Plugin     string `json:"plugin"`      // 可选插件
	PluginOpts string `json:"plugin-opts"` // 插件参数
}

func ParseSSLink(link string) (*SsConfig, error) {
	raw := strings.TrimPrefix(link, "ss://")

	var name string
	if idx := strings.Index(raw, "#"); idx != -1 {
		name, _ = url.QueryUnescape(raw[idx+1:])
		raw = raw[:idx] // 保留 base64 编码部分
	}

	// 尝试 Base64 解码
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ss link: %w", err)
		}
	}

	raw = string(decoded)

	// 解析 METHOD:PASSWORD@SERVER:PORT
	atIdx := strings.LastIndex(raw, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid ss link, missing '@'")
	}

	userInfo := raw[:atIdx]
	serverPart := raw[atIdx+1:]

	colonIdx := strings.LastIndex(serverPart, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid ss link, missing port")
	}

	port, err := strconv.Atoi(serverPart[colonIdx+1:])
	if err != nil {
		return nil, err
	}

	methodPass := strings.SplitN(userInfo, ":", 2)
	if len(methodPass) != 2 {
		return nil, fmt.Errorf("invalid ss link, missing method or password")
	}

	return &SsConfig{
		Name:     name,
		Server:   serverPart[:colonIdx],
		Port:     port,
		Method:   methodPass[0],
		Password: methodPass[1],
	}, nil
}

func ParseVmessLink(link string) (*VmessConfig, error) {
	raw := strings.TrimPrefix(link, "vmess://")

	// 补全 Base64 填充
	padding := len(raw) % 4
	if padding != 0 {
		raw += strings.Repeat("=", 4-padding)
	}

	// 尝试 RawURLEncoding
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		// 再尝试 StdEncoding
		decoded, err = base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess link: %w", err)
		}
	}

	// 解析 JSON
	var cfg VmessConfig
	if err := json.Unmarshal(decoded, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vmess json: %w", err)
	}
	return &cfg, nil
}
