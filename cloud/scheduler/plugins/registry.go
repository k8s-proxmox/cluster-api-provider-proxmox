package plugins

import (
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

func NewVMIDPlugins() []framework.VMIDPlugin {
	return []framework.VMIDPlugin{&nextid.NextID{}}
}
