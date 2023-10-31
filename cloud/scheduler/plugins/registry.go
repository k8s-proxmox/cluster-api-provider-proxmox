package plugins

import (
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/idrange"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/noderesource"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/overcommit"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/regex"
)

func NewNodeFilterPlugins() []framework.NodeFilterPlugin {
	return []framework.NodeFilterPlugin{
		&nodename.NodeName{},
		&overcommit.CPUOvercommit{},
		&overcommit.MemoryOvercommit{},
		&regex.NodeRegex{},
	}
}

func NewNodeScorePlugins() []framework.NodeScorePlugin {
	return []framework.NodeScorePlugin{
		// &random.Random{},
		&noderesource.NodeResource{},
	}
}

func NewVMIDPlugins() []framework.VMIDPlugin {
	return []framework.VMIDPlugin{
		&idrange.Range{},
		&regex.Regex{},
	}
}
