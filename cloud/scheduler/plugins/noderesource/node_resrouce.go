package noderesource

import (
	"context"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
	"github.com/sp-yduck/proxmox-go/api"
)

type NodeResource struct{}

var _ framework.NodeScorePlugin = &NodeResource{}

const (
	Name = names.NodeResource
)

func (pl *NodeResource) Name() string {
	return Name
}

// score = 1/(cpu/maxcpu * mem/maxmem)
func (pl *NodeResource) Score(ctx context.Context, state *framework.CycleState, _ api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	cpu := nodeInfo.Node().Cpu
	maxCPU := nodeInfo.Node().MaxCpu
	mem := nodeInfo.Node().Mem
	maxMem := nodeInfo.Node().MaxMem
	u := cpu / float32(maxCPU) * float32(mem/maxMem)
	score := int64(1 / u)
	return score, nil
}
