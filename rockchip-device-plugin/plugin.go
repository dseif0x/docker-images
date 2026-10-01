package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const healthInterval = 30 * time.Second

var _ dpm.PluginInterface = &Plugin{}

// Plugin advertises one Rockchip resource. The hardware blocks are shared by
// the kernel drivers, so the resource is advertised as `replicas` identical
// slots that all map to the same device nodes.
type Plugin struct {
	resource Resource
	replicas int
}

func (p *Plugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (p *Plugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *Plugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	health := p.health()
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: p.devices(health)}); err != nil {
		glog.Errorf("%s: failed to send ListAndWatch response: %v", p.resource.Name, err)
		return err
	}

	ticker := time.NewTicker(healthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.Context().Done():
			return nil
		case <-ticker.C:
			h := p.health()
			if h == health {
				continue
			}
			health = h
			glog.Infof("%s: health changed to %s", p.resource.Name, health)
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: p.devices(health)}); err != nil {
				glog.Errorf("%s: failed to send ListAndWatch response: %v", p.resource.Name, err)
				return err
			}
		}
	}
}

func (p *Plugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse

	for _, req := range r.ContainerRequests {
		glog.Infof("%s: allocating %v", p.resource.Name, req.DevicesIDs)

		// Every slot maps to the same nodes, so hand them out once per container.
		car := pluginapi.ContainerAllocateResponse{}
		for _, path := range p.resource.Devices {
			car.Devices = append(car.Devices, &pluginapi.DeviceSpec{
				ContainerPath: path,
				HostPath:      path,
				Permissions:   "rw",
			})
		}
		for _, path := range p.resource.Mounts {
			car.Mounts = append(car.Mounts, &pluginapi.Mount{
				ContainerPath: path,
				HostPath:      path,
				ReadOnly:      true,
			})
		}

		response.ContainerResponses = append(response.ContainerResponses, &car)
	}

	return &response, nil
}

func (p *Plugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (p *Plugin) devices(health string) []*pluginapi.Device {
	devices := make([]*pluginapi.Device, p.replicas)
	for i := range devices {
		devices[i] = &pluginapi.Device{
			ID:     fmt.Sprintf("%s-%d", p.resource.Name, i),
			Health: health,
		}
	}
	return devices
}

// health reports Unhealthy if any device node disappeared (e.g. driver unloaded).
func (p *Plugin) health() string {
	for _, path := range p.resource.Devices {
		if _, err := os.Stat(path); err != nil {
			return pluginapi.Unhealthy
		}
	}
	return pluginapi.Healthy
}
