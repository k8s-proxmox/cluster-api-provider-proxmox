package idrange

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/k8s-proxmox/proxmox-go/api"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
)

type Range struct{}

var _ framework.VMIDPlugin = &Range{}

const (
	Name         = names.Range
	VMIDRangeKey = "vmid.qemu-scheduler/range"
)

func (pl *Range) Name() string {
	return Name
}

func (pl *Range) PluginKey() framework.CtxKey {
	return framework.CtxKey(VMIDRangeKey)
}

// select minimum id being not used in specified range
// range is specified in ctx value (key=vmid.qemu-scheduler/range)
func (pl *Range) Select(ctx context.Context, state *framework.CycleState, _ api.VirtualMachineCreateOptions, nextid int, usedID map[int]bool) (int, error) {
	start, end, err := findVMIDRange(ctx)
	if err != nil {
		state.SetMessage(pl.Name(), "no idrange is specified, use nextid.")
		return nextid, nil
	}
	for i := start; i <= end; i++ {
		_, used := usedID[i]
		if !used {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no available vmid in range %d-%d", start, end)
}

// specify available vmid as range
// example: vmid.qemu-scheduler/range=start-end
func findVMIDRange(ctx context.Context) (int, int, error) {
	value := ctx.Value(framework.CtxKey(VMIDRangeKey))
	if value == nil {
		return 0, 0, fmt.Errorf("no vmid range is specified")
	}
	rangeStrs := strings.Split(fmt.Sprintf("%s", value), "-")
	start, err := strconv.Atoi(rangeStrs[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range is specified: %w", err)
	}
	end, err := strconv.Atoi(rangeStrs[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range is specified: %w", err)
	}
	return start, end, nil
}
