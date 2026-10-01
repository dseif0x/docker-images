package main

import (
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
)

var _ dpm.ListerInterface = &Lister{}

type Lister struct {
	resources map[string]Resource
	replicas  map[string]int
}

func (l *Lister) GetResourceNamespace() string {
	return "rockchip.com"
}

// Discover publishes the statically discovered resources once.
func (l *Lister) Discover(pluginListChan chan dpm.PluginNameList) {
	var names dpm.PluginNameList
	for name := range l.resources {
		names = append(names, name)
	}
	pluginListChan <- names
}

func (l *Lister) NewPlugin(name string) dpm.PluginInterface {
	return &Plugin{resource: l.resources[name], replicas: l.replicas[name]}
}
