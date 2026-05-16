package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type NodePingCache struct {
	UpdatedAt string                    `json:"updated_at"`
	Results   map[string]NodePingResult `json:"results"`
}

type NodePingResult struct {
	Tag        string `json:"tag"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Status     string `json:"status"`
	LatencyMS  int64  `json:"latency_ms"`
	Region     string `json:"region"`
	IP         string `json:"ip,omitempty"`
	Error      string `json:"error,omitempty"`
	CheckedAt  string `json:"checked_at"`
}

func DefaultPingCachePath(configPath string) string {
	dir := filepath.Dir(configPath)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, ".proxyctl-ping.json")
}

func PingCacheKey(outbound Outbound, index int) string {
	if outbound.Tag != "" {
		return outbound.Tag
	}
	return fmt.Sprintf("%s:%d#%d", outbound.Server, outbound.ServerPort, index)
}

func LoadPingCache(path string) (NodePingCache, error) {
	cache := NodePingCache{
		Results: map[string]NodePingResult{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return cache, err
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return cache, err
	}
	if cache.Results == nil {
		cache.Results = map[string]NodePingResult{}
	}
	return cache, nil
}

func SavePingCache(path string, cache NodePingCache) error {
	if cache.Results == nil {
		cache.Results = map[string]NodePingResult{}
	}
	cache.UpdatedAt = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(cache, "", "  ")
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

func MatchPingResult(outbound Outbound, result NodePingResult) bool {
	return outbound.Server == result.Server && outbound.ServerPort == result.ServerPort
}

func FormatLatency(result NodePingResult) string {
	if result.Status != "ok" {
		return "-"
	}
	return strconv.FormatInt(result.LatencyMS, 10) + "ms"
}

func FormatRegion(region string) string {
	region = strings.TrimSpace(region)
	if region == "" {
		return "unknown"
	}
	if idx := strings.Index(region, "/"); idx != -1 {
		region = strings.TrimSpace(region[:idx])
	}
	if region == "" {
		return "unknown"
	}
	return region
}
