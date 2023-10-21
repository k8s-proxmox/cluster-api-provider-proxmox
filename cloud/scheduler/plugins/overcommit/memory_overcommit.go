package overcommit

import (
	"context"

	"github.com/sp-yduck/proxmox-go/api"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
)

type MemoryOvercommit struct{}

var _ framework.NodeFilterPlugin = &MemoryOvercommit{}

const (
	MemoryOvercommitName         = names.MemoryOvercommit
	defaultMemoryOvercommitRatio = 1
)

func (pl *MemoryOvercommit) Name() string {
	return MemoryOvercommitName
}

// filter by memory overcommit ratio
func (pl *MemoryOvercommit) Filter(ctx context.Context, _ *framework.CycleState, config api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) *framework.Status {
	mem := sumMems(nodeInfo.QEMUs())
	maxMem := nodeInfo.Node().MaxMem
	ratio := float32(mem+1024*1024*config.Memory) / float32(maxMem)
	if ratio >= defaultMemoryOvercommitRatio {
		status := framework.NewStatus()
		status.SetCode(1)
		return status
	}
	return &framework.Status{}
}

func sumMems(qemus []*api.VirtualMachine) int {
	var result int
	for _, q := range qemus {
		result += q.MaxMem
	}
	return result
}
