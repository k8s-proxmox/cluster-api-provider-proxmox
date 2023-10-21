package regex

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sp-yduck/proxmox-go/api"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/names"
)

type NodeRegex struct{}

var _ framework.NodeFilterPlugin = &NodeRegex{}

const (
	NodeRegexName = names.NodeRegex
	NodeRegexKey  = "node.qemu-scheduler/regex"
)

func (pl *NodeRegex) Name() string {
	return NodeRegexName
}

// regex is specified in ctx value (key=node.qemu-scheduler/regex)
func (pl *NodeRegex) Filter(ctx context.Context, _ *framework.CycleState, config api.VirtualMachineCreateOptions, nodeInfo *framework.NodeInfo) *framework.Status {
	reg, err := findNodeRegex(ctx)
	if err != nil {

		return &framework.Status{}
	}
	if !reg.MatchString(nodeInfo.Node().Node) {
		status := framework.NewStatus()
		status.SetCode(1)
		return status
	}
	return &framework.Status{}
}

// specify available node name as regex
// example: node.qemu-scheduler/regex=node[0-9]+
func findNodeRegex(ctx context.Context) (*regexp.Regexp, error) {
	value := ctx.Value(framework.CtxKey(NodeRegexKey))
	if value == nil {
		return nil, fmt.Errorf("no node name regex is specified")
	}
	reg, err := regexp.Compile(fmt.Sprintf("%s", value))
	if err != nil {
		return nil, err
	}
	return reg, nil
}
