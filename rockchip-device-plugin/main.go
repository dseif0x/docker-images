package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

func main() {
	npuReplicas := flag.Int("npu-replicas", 1, "number of pods that may share the NPU")
	vpuReplicas := flag.Int("vpu-replicas", 1, "number of pods that may share the VPU")
	list := flag.Bool("list", false, "print discovered resources and exit")
	// glog writes to files in /tmp unless told otherwise.
	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	if platform, err := GetHardwarePlatform(); err != nil {
		glog.Warningf("failed to read hardware platform: %v", err)
	} else {
		glog.Infof("Hardware platform: %s", platform)
	}

	resources := make(map[string]Resource)
	for _, r := range Discover() {
		glog.Infof("rockchip.com/%s: devices=%v mounts=%v", r.Name, r.Devices, r.Mounts)
		resources[r.Name] = r
	}

	if *list {
		for _, r := range resources {
			fmt.Printf("rockchip.com/%s\n  devices: %v\n  mounts: %v\n", r.Name, r.Devices, r.Mounts)
		}
		return
	}

	lister := &Lister{
		resources: resources,
		replicas:  map[string]int{"npu": max(*npuReplicas, 1), "vpu": max(*vpuReplicas, 1)},
	}
	dpm.NewManager(lister).Run()
}
