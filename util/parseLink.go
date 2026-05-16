package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
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

func decodeBase64(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "=")
	if decoded, err := base64.RawURLEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}

	padding := len(raw) % 4
	if padding != 0 {
		raw += strings.Repeat("=", 4-padding)
	}
	if decoded, err := base64.URLEncoding.DecodeString(raw); err == nil {
		return decoded, nil
	}
	return base64.StdEncoding.DecodeString(raw)
}

func splitHostPort(serverPart string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(serverPart)
	if err != nil {
		colonIdx := strings.LastIndex(serverPart, ":")
		if colonIdx == -1 {
			return "", 0, fmt.Errorf("invalid server address, missing port")
		}
		host = serverPart[:colonIdx]
		portStr = serverPart[colonIdx+1:]
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %w", err)
	}
	return strings.Trim(host, "[]"), port, nil
}

func ParseSSLink(link string) (*SsConfig, error) {
	raw := strings.TrimPrefix(link, "ss://")

	var name string
	if idx := strings.Index(raw, "#"); idx != -1 {
		name, _ = url.QueryUnescape(raw[idx+1:])
		raw = raw[:idx]
	}

	plugin := ""
	pluginOpts := ""
	if idx := strings.Index(raw, "?"); idx != -1 {
		values, _ := url.ParseQuery(raw[idx+1:])
		plugin = values.Get("plugin")
		pluginOpts = values.Get("plugin-opts")
		raw = raw[:idx]
	}

	// SIP002: ss://BASE64(method:password)@server:port
	if atIdx := strings.LastIndex(raw, "@"); atIdx != -1 {
		userInfo := raw[:atIdx]
		serverPart := raw[atIdx+1:]

		decodedUserInfo, err := decodeBase64(userInfo)
		if err == nil {
			userInfo = string(decodedUserInfo)
		} else {
			userInfo, _ = url.QueryUnescape(userInfo)
		}

		methodPass := strings.SplitN(userInfo, ":", 2)
		if len(methodPass) != 2 {
			return nil, fmt.Errorf("invalid ss link, missing method or password")
		}

		server, port, err := splitHostPort(serverPart)
		if err != nil {
			return nil, err
		}

		return &SsConfig{
			Name:       name,
			Server:     server,
			Port:       port,
			Method:     methodPass[0],
			Password:   methodPass[1],
			Plugin:     plugin,
			PluginOpts: pluginOpts,
		}, nil
	}

	// Legacy form: ss://BASE64(method:password@server:port)
	decoded, err := decodeBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ss link: %w", err)
	}

	raw = string(decoded)

	atIdx := strings.LastIndex(raw, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid ss link, missing '@'")
	}
	userInfo := raw[:atIdx]
	serverPart := raw[atIdx+1:]

	server, port, err := splitHostPort(serverPart)
	if err != nil {
		return nil, err
	}

	methodPass := strings.SplitN(userInfo, ":", 2)
	if len(methodPass) != 2 {
		return nil, fmt.Errorf("invalid ss link, missing method or password")
	}

	return &SsConfig{
		Name:       name,
		Server:     server,
		Port:       port,
		Method:     methodPass[0],
		Password:   methodPass[1],
		Plugin:     plugin,
		PluginOpts: pluginOpts,
	}, nil
}

func ParseVmessLink(link string) (*VmessConfig, error) {
	raw := strings.TrimPrefix(link, "vmess://")

	decoded, err := decodeBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vmess link: %w", err)
	}

	// 解析 JSON
	var cfg VmessConfig
	if err := json.Unmarshal(decoded, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vmess json: %w", err)
	}
	return &cfg, nil
}
