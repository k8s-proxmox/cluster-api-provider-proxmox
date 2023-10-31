package framework

import (
	"context"

	"github.com/sp-yduck/proxmox-go/api"
)

type Plugin interface {
	// return plugin name
	Name() string
}

type NodeFilterPlugin interface {
	Plugin
	Filter(ctx context.Context, state *CycleState, config api.VirtualMachineCreateOptions, nodeInfo *NodeInfo) *Status
}

type NodeScorePlugin interface {
	Plugin
	Score(ctx context.Context, state *CycleState, config api.VirtualMachineCreateOptions, nodeInfo *NodeInfo) (int64, *Status)
}

type VMIDPlugin interface {
	Plugin
	PluginKey() CtxKey
	Select(ctx context.Context, state *CycleState, config api.VirtualMachineCreateOptions, nextid int, usedID map[int]bool) (int, error)
}
