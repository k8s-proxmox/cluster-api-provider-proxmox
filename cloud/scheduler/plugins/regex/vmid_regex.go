package regex

import (
	"context"
	"fmt"
	"regexp"

	"github.com/k8s-proxmox/proxmox-go/api"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
)

type Regex struct{}

var _ framework.VMIDPlugin = &Regex{}

const (
	Name         = names.Regex
	VMIDRegexKey = "vmid.qemu-scheduler/regex"
)

func (pl *Regex) Name() string {
	return Name
}

func (pl *Regex) PluginKey() framework.CtxKey {
	return framework.CtxKey(VMIDRegexKey)
}

// select minimum id being not used and matching regex
// regex is specified in ctx value (key=vmid.qemu-scheduler/regex)
func (pl *Regex) Select(ctx context.Context, state *framework.CycleState, _ api.VirtualMachineCreateOptions, nextid int, usedID map[int]bool) (int, error) {
	idregex, err := findVMIDRegex(ctx)
	if err != nil {
		state.SetMessage(pl.Name(), "no idregex is specified, use nextid.")
		return nextid, nil
	}
	for i := nextid; i < 1000000000; i++ {
		_, used := usedID[i]
		if idregex.MatchString(fmt.Sprintf("%d", i)) && !used {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no available vmid")
}

// specify available vmid as regex
// example: vmid.qemu-scheduler/regex=(12[0-9]|130)
func findVMIDRegex(ctx context.Context) (*regexp.Regexp, error) {
	value := ctx.Value(framework.CtxKey(VMIDRegexKey))
	if value == nil {
		return nil, fmt.Errorf("no vmid regex is specified")
	}
	reg, err := regexp.Compile(fmt.Sprintf("%s", value))
	if err != nil {
		return nil, err
	}
	return reg, nil
}
