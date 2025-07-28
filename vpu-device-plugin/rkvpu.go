package main

import (
	"fmt"
	"os"
	"strings"
)

var DevicePaths = []string{"/dev/dri", "/dev/dma_heap", "/dev/mali0", "/dev/rga", "/dev/mpp_service"}

const (
	compatiblePath = "/proc/device-tree/compatible"
)

var platforms = []string{"rk3588", "rk3576", "rk3566", "rk3568", "rk3562", "rv1103", "rv1106", "rv1103b", "rv1106b", "rk2118"}

func GetHardwarePlatform() (string, error) {
	compatible, err := os.ReadFile(compatiblePath)
	if err != nil {
		return "", fmt.Errorf("failed to read hardware platform: %v", err)
	}

	compatibleStr := strings.Split(string(compatible), ",")
	// Check if any of the compatible strings contain a supported platform
	for _, comp := range compatibleStr {
		for _, platform := range platforms {
			if strings.Contains(comp, platform) {
				return platform, nil
			}
		}
	}

	return compatibleStr[0], fmt.Errorf("platforms %v, not supported", strings.Join(compatibleStr, ", "))
}
