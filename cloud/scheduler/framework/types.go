package framework

import (
	"context"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
)

type Status struct {
	code         int
	reasons      []string
	err          error
	failedPlugin string
}

func NewStatus() *Status {
	return &Status{code: 0}
}

func (s *Status) Code() int {
	return s.code
}

func (s *Status) SetCode(code int) {
	s.code = code
}

func (s *Status) Reasons() []string {
	if s.err != nil {
		return append([]string{s.err.Error()}, s.reasons...)
	}
	return s.reasons
}

func (s *Status) FailedPlugin() string {
	return s.failedPlugin
}

func (s *Status) SetFailedPlugin(name string) {
	s.failedPlugin = name
}

func (s *Status) IsSuccess() bool {
	return s.code == 0
}

func (s *Status) Error() error {
	return s.err
}

// NodeInfo is node level aggregated information
type NodeInfo struct {
	node *api.Node

	// qemus assigned to the node
	qemus []*api.VirtualMachine
}

func GetNodeInfoList(ctx context.Context, client *proxmox.Service) ([]*NodeInfo, error) {
	nodes, err := client.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	nodeInfos := []*NodeInfo{}
	for _, node := range nodes {
		qemus, err := client.RESTClient().GetVirtualMachines(ctx, node.Node)
		if err != nil {
			return nil, err
		}
		nodeInfos = append(nodeInfos, &NodeInfo{node: node, qemus: qemus})
	}
	return nodeInfos, nil
}

func (n NodeInfo) Node() *api.Node {
	return n.node
}

// NodeScoreList declares a list of nodes and their scores.
type NodeScoreList []NodeScore

// NodeScore is a struct with node name and score.
type NodeScore struct {
	Name  string
	Score int64
}
