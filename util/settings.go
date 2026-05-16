package util

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProxyctlSettings struct {
	Inbound InboundSettings `json:"inbound"`
}

type InboundSettings struct {
	Mode      string `json:"mode"`
	Listen    string `json:"listen"`
	SocksPort int    `json:"socks_port"`
	HTTPPort  int    `json:"http_port"`
	MixedPort int    `json:"mixed_port"`
}

func DefaultSettingsPath(configPath string) string {
	dir := filepath.Dir(configPath)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, ".proxyctl-settings.json")
}

func DefaultProxyctlSettings() ProxyctlSettings {
	opts := DefaultGenerateConfigOptions("", 0)
	return ProxyctlSettings{
		Inbound: InboundSettings{
			Mode:      opts.InboundMode,
			Listen:    opts.Listen,
			SocksPort: opts.SocksPort,
			HTTPPort:  opts.HTTPPort,
			MixedPort: opts.MixedPort,
		},
	}
}

func (settings ProxyctlSettings) Normalize() ProxyctlSettings {
	defaults := DefaultProxyctlSettings()
	if settings.Inbound.Mode == "" {
		settings.Inbound.Mode = defaults.Inbound.Mode
	}
	if settings.Inbound.Listen == "" {
		settings.Inbound.Listen = defaults.Inbound.Listen
	}
	if settings.Inbound.SocksPort == 0 {
		settings.Inbound.SocksPort = defaults.Inbound.SocksPort
	}
	if settings.Inbound.HTTPPort == 0 {
		settings.Inbound.HTTPPort = defaults.Inbound.HTTPPort
	}
	if settings.Inbound.MixedPort == 0 {
		settings.Inbound.MixedPort = defaults.Inbound.MixedPort
	}
	return settings
}

func LoadProxyctlSettings(path string) (ProxyctlSettings, error) {
	settings := DefaultProxyctlSettings()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, err
	}
	return settings.Normalize(), nil
}

func SaveProxyctlSettings(path string, settings ProxyctlSettings) error {
	settings = settings.Normalize()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}

func (settings ProxyctlSettings) GenerateConfigOptions(outPath string, defaultIndex int) GenerateConfigOptions {
	settings = settings.Normalize()
	return GenerateConfigOptions{
		OutPath:      outPath,
		DefaultIndex: defaultIndex,
		InboundMode:  settings.Inbound.Mode,
		Listen:       settings.Inbound.Listen,
		SocksPort:    settings.Inbound.SocksPort,
		HTTPPort:     settings.Inbound.HTTPPort,
		MixedPort:    settings.Inbound.MixedPort,
	}
}
