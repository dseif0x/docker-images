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

	go func() {
		if _, err := os.Stat(devicePath); os.IsNotExist(err) {
			glog.Errorf("Device not found: %s", devicePath)
			return
		}

		lister.ResUpdateChan <- dpm.PluginNameList{"gpu"}
	}()

	manager.Run()
}
