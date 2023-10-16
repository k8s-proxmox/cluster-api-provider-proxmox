package plugins

import (
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/node/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/node/random"
)

func NewNodeFilterPlugins() []framework.NodeFilterPlugin {
	return []framework.NodeFilterPlugin{&nodename.NodeName{}}
}

func NewNodeScorePlugins() []framework.NodeScorePlugin {
	return []framework.NodeScorePlugin{&random.Random{}}
}
