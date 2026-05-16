package util

import (
	"fmt"
	"strconv"
	"strings"
)

func NodeDisplayName(outbound Outbound, index int) string {
	tag := strings.TrimSpace(outbound.Tag)
	if tag == "" {
		return fmt.Sprintf("%s:%d", outbound.Server, outbound.ServerPort)
	}

	parts := strings.SplitN(tag, "-", 3)
	if len(parts) == 3 {
		if _, err := strconv.Atoi(parts[1]); err == nil && parts[2] != "" {
			return parts[2]
		}
	}

	return tag
}
