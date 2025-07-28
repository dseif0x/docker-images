package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

func main() {
	defer glog.Flush()

	lister := NewLister()
	manager := dpm.NewManager(lister)

	platform, err := GetHardwarePlatform()
	if err != nil {
		glog.Errorf("failed to get hardware platform: %v", err)
		manager.Run()
		return
	}

	glog.Infof("Hardware platform: %s", platform)

	go func() {
		for _, devicePath := range DevicePaths {
			if _, err := os.Stat(devicePath); os.IsNotExist(err) {
				glog.Errorf("Device not found: %s", devicePath)
				return
			}
		}

		lister.ResUpdateChan <- dpm.PluginNameList{"vpu"}
	}()

	manager.Run()
}
