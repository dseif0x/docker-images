package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang/glog"
)

// Roots are variables so tests can point discovery at a fake tree.
var (
	sysRoot = "/sys"
	devRoot = "/dev"
)

const compatiblePath = "/proc/device-tree/compatible"

var platforms = []string{"rk3588", "rk3576", "rk3566", "rk3568", "rk3562", "rv1103", "rv1106", "rv1103b", "rv1106b", "rk2118"}

// Resource is one schedulable resource (rockchip.com/<Name>) and the host
// device nodes / mounts handed to every container that is allocated it.
type Resource struct {
	Name    string
	Devices []string // absolute paths under /dev
	Mounts  []string // absolute host paths bind-mounted read-only at the same path
}

// Discover returns the resources present on this node. A resource is only
// returned when all its required device nodes exist.
func Discover() []Resource {
	npuNodes := npuDevices()

	var resources []Resource
	if len(npuNodes) > 0 {
		npu := Resource{Name: "npu", Devices: npuNodes}
		// rknn-toolkit reads the SoC model from the device tree.
		if _, err := os.Stat(compatiblePath); err == nil {
			npu.Mounts = append(npu.Mounts, compatiblePath)
		}
		resources = append(resources, npu)
	} else {
		glog.Info("no NPU found")
	}

	if vpuNodes := vpuDevices(npuNodes); len(vpuNodes) > 0 {
		resources = append(resources, Resource{Name: "vpu", Devices: vpuNodes})
	} else {
		glog.Info("no VPU found (/dev/mpp_service missing)")
	}

	return resources
}

// npuDevices returns the NPU device nodes. The vendor rknpu driver registers
// a DRM card + render node whose minor number depends on probe order, so it is
// looked up by driver name instead of assuming renderD129. Older vendor
// kernels expose /dev/rknpu, mainline (rocket) uses /dev/accel/accelN.
func npuDevices() []string {
	nodes := classNodesByDriver("drm", "dri", func(drv string) bool { return strings.EqualFold(drv, "rknpu") })
	nodes = append(nodes, classNodesByDriver("accel", "accel", func(drv string) bool { return drv == "rocket" })...)
	if exists("rknpu") {
		nodes = append(nodes, "/dev/rknpu")
	}
	return nodes
}

// vpuDevices returns the nodes needed for Rockchip MPP hardware video
// decode/encode plus the helpers ffmpeg-rockchip/Jellyfin/Frigate use with it:
// RGA (scaling/colour conversion), DMA heaps (buffer sharing), the Mali GPU
// and the non-NPU DRM nodes. NPU nodes are excluded so a VPU claim does not
// grant NPU access.
func vpuDevices(npuNodes []string) []string {
	if !exists("mpp_service") {
		return nil
	}

	nodes := []string{"/dev/mpp_service"}
	for _, name := range []string{"rga", "mali0"} {
		if exists(name) {
			nodes = append(nodes, "/dev/"+name)
		}
	}
	nodes = append(nodes, dirNodes("dma_heap")...)

	npu := make(map[string]bool, len(npuNodes))
	for _, n := range npuNodes {
		npu[n] = true
	}
	for _, n := range dirNodes("dri") {
		if !npu[n] {
			nodes = append(nodes, n)
		}
	}

	return nodes
}

// classNodesByDriver returns /dev/<devDir>/<name> for every entry of
// /sys/class/<class> whose bound driver matches and whose node exists.
func classNodesByDriver(class, devDir string, match func(string) bool) []string {
	entries, err := os.ReadDir(filepath.Join(sysRoot, "class", class))
	if err != nil {
		return nil
	}

	var nodes []string
	for _, e := range entries {
		// Skip connectors such as card0-HDMI-A-1.
		if strings.Contains(e.Name(), "-") {
			continue
		}
		drv, err := filepath.EvalSymlinks(filepath.Join(sysRoot, "class", class, e.Name(), "device", "driver"))
		if err != nil || !match(filepath.Base(drv)) {
			continue
		}
		rel := filepath.Join(devDir, e.Name())
		if !exists(rel) {
			glog.Warningf("%s is bound to %s but /dev/%s does not exist", e.Name(), filepath.Base(drv), rel)
			continue
		}
		nodes = append(nodes, "/dev/"+rel)
	}
	sort.Strings(nodes)
	return nodes
}

// dirNodes returns the non-directory entries of /dev/<dir> (e.g. by-path is skipped).
func dirNodes(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(devRoot, dir))
	if err != nil {
		return nil
	}

	var nodes []string
	for _, e := range entries {
		if !e.IsDir() {
			nodes = append(nodes, "/dev/"+dir+"/"+e.Name())
		}
	}
	return nodes
}

// exists reports whether /dev/<rel> exists.
func exists(rel string) bool {
	_, err := os.Stat(filepath.Join(devRoot, rel))
	return err == nil
}

func GetHardwarePlatform() (string, error) {
	compatible, err := os.ReadFile(compatiblePath)
	if err != nil {
		return "", err
	}

	// The file is a list of NUL-separated "vendor,model" strings.
	for _, comp := range strings.Split(string(compatible), "\x00") {
		for _, platform := range platforms {
			if strings.Contains(comp, platform) {
				return platform, nil
			}
		}
	}

	return strings.ReplaceAll(string(compatible), "\x00", " "), nil
}
