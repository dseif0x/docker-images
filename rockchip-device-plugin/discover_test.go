package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fakeTree builds an RK3588-like /sys and /dev (vendor kernel) under a temp dir.
func fakeTree(t *testing.T, withMpp bool) {
	t.Helper()
	root := t.TempDir()
	sysRoot = filepath.Join(root, "sys")
	devRoot = filepath.Join(root, "dev")

	mkdir := func(p string) {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	touch := func(p string) {
		mkdir(filepath.Dir(p))
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	link := func(target, name string) {
		mkdir(filepath.Dir(name))
		if err := os.Symlink(target, name); err != nil {
			t.Fatal(err)
		}
	}

	drivers := filepath.Join(sysRoot, "bus/platform/drivers")
	mkdir(filepath.Join(drivers, "rockchip-drm"))
	mkdir(filepath.Join(drivers, "RKNPU"))
	display := filepath.Join(sysRoot, "devices/platform/display-subsystem")
	npu := filepath.Join(sysRoot, "devices/platform/fdab0000.npu")
	link(filepath.Join(drivers, "rockchip-drm"), filepath.Join(display, "driver"))
	link(filepath.Join(drivers, "RKNPU"), filepath.Join(npu, "driver"))

	class := filepath.Join(sysRoot, "class/drm")
	for name, dev := range map[string]string{
		"card0": display, "card0-HDMI-A-1": display, "renderD128": display,
		"card1": npu, "renderD129": npu,
	} {
		link(dev, filepath.Join(class, name, "device"))
	}

	for _, n := range []string{"dri/card0", "dri/card1", "dri/renderD128", "dri/renderD129", "rga", "mali0", "dma_heap/system", "dma_heap/cma"} {
		touch(filepath.Join(devRoot, n))
	}
	mkdir(filepath.Join(devRoot, "dri/by-path"))
	if withMpp {
		touch(filepath.Join(devRoot, "mpp_service"))
	}
}

func TestDiscover(t *testing.T) {
	fakeTree(t, true)

	got := map[string][]string{}
	for _, r := range Discover() {
		got[r.Name] = r.Devices
	}

	want := map[string][]string{
		"npu": {"/dev/dri/card1", "/dev/dri/renderD129"},
		"vpu": {"/dev/mpp_service", "/dev/rga", "/dev/mali0", "/dev/dma_heap/cma", "/dev/dma_heap/system", "/dev/dri/card0", "/dev/dri/renderD128"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Discover() =\n%v\nwant\n%v", got, want)
	}
}

func TestDiscoverWithoutMpp(t *testing.T) {
	fakeTree(t, false)

	resources := Discover()
	if len(resources) != 1 || resources[0].Name != "npu" {
		t.Errorf("Discover() = %+v, want only npu", resources)
	}
}
