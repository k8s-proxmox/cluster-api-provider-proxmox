package plugins

import (
	"fmt"

	"github.com/sp-yduck/proxmox-go/proxmox"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nextid"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/random"
)

func VMIDPlugins(client *proxmox.Service) map[string]framework.VMIDPlugin {
	return map[string]framework.VMIDPlugin{
		names.NextID: nextid.New(client),
	}
}

func NewNodeFilterPlugins() []framework.NodeFilterPlugin {
	return []framework.NodeFilterPlugin{&nodename.NodeName{}}
}

func NewNodeScorePlugins() []framework.NodeScorePlugin {
	return []framework.NodeScorePlugin{&random.Random{}}
}

func NewVMIDPlugin(client *proxmox.Service, name string) (framework.VMIDPlugin, error) {
	plugins := VMIDPlugins(client)
	plugin, ok := plugins[name]
	if !ok {
		return nil, fmt.Errorf("vmid plugin %s not found", name)
	}
	return plugin, nil
}
