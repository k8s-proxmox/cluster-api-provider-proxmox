package overcommit

import (
	"context"

	"github.com/k8s-proxmox/proxmox-go/api"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
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
func (pl *MemoryOvercommit) Filter(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) *framework.Status {
	mem := sumMems(nodeInfo.QEMUs())
	maxMem := nodeInfo.Node().MaxMem
	ratio := float32(mem+1024*1024*config.Memory) / float32(maxMem)
	if ratio >= defaultMemoryOvercommitRatio {
		status := framework.NewStatus()
		status.SetCode(1)
		state.SetMessage(pl.Name(), "exceed memory overcommit ratio")
		return status
	}
	return &framework.Status{}
}

// sum maxmem of all 'running' qemu
func sumMems(qemus []*api.VirtualMachine) int {
	var result int
	for _, q := range qemus {
		if q.Status == api.ProcessStatusRunning {
			result += q.MaxMem
		}
	}
	return result
}
