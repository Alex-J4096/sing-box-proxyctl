package util

import (
	"os"
	"path/filepath"
)

const appDirName = "sing-box-proxyctl"

func DefaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || home == "" {
			return "."
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, appDirName)
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}
