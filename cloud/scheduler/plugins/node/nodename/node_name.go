package nodename

import (
	"context"

	"github.com/sp-yduck/proxmox-go/api"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/node/names"
)

type NodeName struct{}

var _ framework.NodeFilterPlugin = &NodeName{}

const (
	Name      = names.NodeName
	ErrReason = "node didn't match the requested node name"
)

func (pl *NodeName) Name() string {
	return Name
}

func (pl *NodeName) Filter(ctx context.Context, _ *framework.CycleState, config api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) *framework.Status {
	if !Fits(config, nodeInfo) {
		status := framework.NewStatus()
		status.SetCode(1)
		return status
	}
	return &framework.Status{}
}

// return true if config.Node is empty or match with node name
func Fits(config api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) bool {
	return config.Node == "" || config.Node == nodeInfo.Node().Node
}
