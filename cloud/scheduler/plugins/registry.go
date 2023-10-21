package plugins

import (
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/idrange"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/random"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/regex"
)

func NewNodeFilterPlugins() []framework.NodeFilterPlugin {
	return []framework.NodeFilterPlugin{
		&nodename.NodeName{},
	}
}

func NewNodeScorePlugins() []framework.NodeScorePlugin {
	return []framework.NodeScorePlugin{
		&random.Random{},
	}
}

func NewVMIDPlugins() []framework.VMIDPlugin {
	return []framework.VMIDPlugin{
		&idrange.Range{},
		&regex.Regex{},
	}
}
