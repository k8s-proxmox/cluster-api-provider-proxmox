package nextid

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
)

type NextID struct {
	client *proxmox.Service
}

var _ framework.VMIDPlugin = &NextID{}

const (
	Name         = names.NextID
	VMIDRegexKey = "vmid.qemu-scheduler/regex"
)

func New(client *proxmox.Service) *NextID {
	return &NextID{client: client}
}

func (pl *NextID) Name() string {
	return Name
}

// select minimum id being not used and matching regex
// regex is specified in ctx value (key=vmid.qemu-scheduler/regex)
func (pl *NextID) Select(ctx context.Context, _ *framework.CycleState, _ api.VirtualMachineCreateOptions) (int, error) {
	nextid, err := pl.client.RESTClient().GetNextID(ctx)
	if err != nil {
		return 0, err
	}
	idrange, err := findVMIDRange(ctx)
	if err != nil {
		return nextid, nil
	}
	table, err := pl.usedIDMap(ctx)
	if err != nil {
		return 0, err
	}
	for i := nextid; i < 1000000000; i++ {
		_, used := table[i]
		if idrange.MatchString(fmt.Sprintf("%d", i)) && !used {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no available vmid")
}

// specify available vmid as regex
// example: vmid.qemu-scheduler/regex=(12[0-9]|130)
func findVMIDRange(ctx context.Context) (*regexp.Regexp, error) {
	value := ctx.Value(framework.CtxKey(VMIDRegexKey))
	if value == nil {
		return nil, fmt.Errorf("no vmid range is specified")
	}
	reg, err := regexp.Compile(fmt.Sprintf("%s", value))
	if err != nil {
		return nil, err
	}
	return reg, nil
}

// return map[vmid]bool
func (pl *NextID) usedIDMap(ctx context.Context) (map[int]bool, error) {
	vms, err := pl.client.VirtualMachines(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[int]bool)
	for _, vm := range vms {
		result[vm.VMID] = true
	}
	return result, nil
}
