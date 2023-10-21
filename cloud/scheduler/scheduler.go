package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/queue"
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	// ErrNoNodesAvailable is used to describe the error that no nodes available to schedule qemus.
	ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule qemus")

	// ErrNoVMIDAvailable is used to describe the error that no vmid available to schedule qemus.
	ErrNoVMIDAvailable = fmt.Errorf("no vmid available to schedule qemus")
)

// manager manages schedulers
type Manager struct {
	ctx context.Context

	// params is used for initializing each scheduler
	params SchedulerParams

	// scheduler map
	table map[schedulerID]*Scheduler
}

// return manager with initialized scheduler-table
func NewManager(params SchedulerParams) *Manager {
	table := make(map[schedulerID]*Scheduler)
	return &Manager{ctx: context.Background(), params: params, table: table}
}

// return new/existing scheduler running
func (m *Manager) GetScheduler(client *proxmox.Service) *Scheduler {
	logger := m.params.Logger.WithValues("Name", "qemu-scheduler")
	schedID, err := m.getSchedulerID(client)
	if err != nil {
		// create new scheduler without registering
		// to not make it zombie scheduler set timeout to context
		ctx, cancel := context.WithTimeout(m.ctx, 5*time.Minute)
		sched := m.NewScheduler(client, logger, cancel)
		return sched.WithRun(ctx)
	}
	logger = logger.WithValues("scheduler ID", *schedID)
	sched, ok := m.table[*schedID]
	if !ok {
		// create and register new scheduler
		logger.V(5).Info("registering new scheduler")
		ctx, cancel := context.WithCancel(m.ctx)
		sched := m.NewScheduler(client, logger, cancel)
		m.table[*schedID] = sched
		return sched.WithRun(ctx)
	}
	logger.V(5).Info("using existing scheduler")
	return sched
}

func (m *Manager) NewScheduler(client *proxmox.Service, logger logr.Logger, cancel context.CancelFunc) *Scheduler {
	return &Scheduler{
		client:          client,
		schedulingQueue: queue.New(),

		filterPlugins: plugins.NewNodeFilterPlugins(),
		scorePlugins:  plugins.NewNodeScorePlugins(),
		vmidPlugins:   plugins.NewVMIDPlugins(),

		resultMap: make(map[string]chan framework.CycleState),
		logger:    logger,
		cancel:    cancel,
	}
}

// get scheduler identifier
// (treat ipaddr&fingreprint of node having id=1 as proxmox cluster identifier)
func (m *Manager) getSchedulerID(client *proxmox.Service) (*schedulerID, error) {
	joinConfig, err := client.JoinConfig(context.Background())
	if err != nil {
		return nil, err
	}
	for _, node := range joinConfig.NodeList {
		if node.NodeID == "1" {
			return &schedulerID{IPAddress: node.PVEAddr, Fingreprint: node.PVEFP}, nil
		}
	}
	return nil, fmt.Errorf("no nodes with id=1")
}

type Scheduler struct {
	client          *proxmox.Service
	schedulingQueue *queue.SchedulingQueue

	filterPlugins []framework.NodeFilterPlugin
	scorePlugins  []framework.NodeScorePlugin
	vmidPlugins   []framework.VMIDPlugin

	// to do : cache

	resultMap map[string]chan framework.CycleState
	logger    logr.Logger

	// to stop itself
	cancel context.CancelFunc
}

type SchedulerParams struct {
	Logger logr.Logger
}

type schedulerID struct {
	IPAddress   string
	Fingreprint string
}

// run scheduler
func (s *Scheduler) Run(ctx context.Context) {
	wait.UntilWithContext(ctx, s.ScheduleOne, 0)
}

// run scheduelr and return it
func (s *Scheduler) WithRun(ctx context.Context) *Scheduler {
	go s.Run(ctx)
	return s
}

// stop scheduler
func (s *Scheduler) Stop() {
	s.cancel()
}

func (s *Scheduler) ScheduleOne(ctx context.Context) {
	qemu := s.schedulingQueue.NextQEMU()
	config := qemu.Config()
	qemuCtx := qemu.Context()
	s.logger = s.logger.WithValues("qemu", config.Name)
	s.logger.Info("scheduling qemu")

	state := framework.NewCycleState()
	s.resultMap[config.Name] = make(chan framework.CycleState, 1)
	defer func() { s.resultMap[config.Name] <- state }()

	node, err := s.SelectNode(qemuCtx, *config)
	if err != nil {
		state.UpdateState(true, err, nil)
		return
	}

	// to do: do this parallel with SelectNode
	vmid, err := s.SelectVMID(qemuCtx, *config)
	if err != nil {
		state.UpdateState(true, err, nil)
		return
	}

	vm, err := s.client.CreateVirtualMachine(ctx, node, vmid, *config)
	if err != nil {
		state.UpdateState(true, err, nil)
		return
	}

	result := framework.NewSchedulerResult(vmid, node, vm)
	state.UpdateState(true, nil, &result)
}

// return status
func (s *Scheduler) WaitStatus(ctx context.Context, config *api.VirtualMachineCreateOptions) (framework.CycleState, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	var done chan framework.CycleState
	ok := false
	for !ok {
		done, ok = s.resultMap[config.Name]
		if !ok {
			time.Sleep(100 * time.Millisecond)
		}
	}
	select {
	case state := <-done:
		return state, nil
	case <-ctx.Done():
		return framework.CycleState{}, fmt.Errorf("exceed timeout deadline")
	}
}

// create new qemu with given spec and context
func (s *Scheduler) CreateQEMU(ctx context.Context, config *api.VirtualMachineCreateOptions) (framework.SchedulerResult, error) {
	s.schedulingQueue.Add(ctx, config)
	status := framework.NewCycleState()
	// to do : timeout
	for !status.IsCompleted() {
		var err error
		status, err = s.WaitStatus(ctx, config)
		if err != nil {
			return status.Result(), err
		}
	}
	if status.Error() != nil {
		return status.Result(), status.Error()
	}
	return status.Result(), nil
}

func (s *Scheduler) SelectNode(ctx context.Context, config api.VirtualMachineCreateOptions) (string, error) {
	s.logger.Info("finding proxmox node matching qemu")
	nodes, err := s.client.Nodes(ctx)
	if err != nil {
		return "", err
	}

	// filter
	nodelist, _ := s.RunFilterPlugins(ctx, nil, config, nodes)
	if len(nodelist) == 0 {
		return "", ErrNoNodesAvailable
	}
	if len(nodelist) == 1 {
		return nodelist[0].Node, nil
	}

	// score
	scorelist, status := s.RunScorePlugins(ctx, nil, config, nodelist)
	if !status.IsSuccess() {
		s.logger.Error(status.Error(), "scoring failed")
	}
	selectedNode, err := selectHighestScoreNode(scorelist)
	if err != nil {
		return "", err
	}
	s.logger.Info(fmt.Sprintf("proxmox node %s was selected for vm %s", selectedNode, config.Name))
	return selectedNode, nil
}

func (s *Scheduler) SelectVMID(ctx context.Context, config api.VirtualMachineCreateOptions) (int, error) {
	s.logger.Info("finding proxmox vmid to be assigned to qemu")
	if config.VMID != nil {
		return *config.VMID, nil
	}

	nextid, err := s.client.NextID(ctx)
	if err != nil {
		return 0, err
	}

	qemus, err := s.client.VirtualMachines(ctx)
	if err != nil {
		return 0, err
	}

	return s.RunVMIDPlugins(ctx, nil, config, nextid, qemus)
}

func (s *Scheduler) RunFilterPlugins(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nodes []*api.Node) ([]*api.Node, error) {
	s.logger.Info("filtering proxmox node")
	feasibleNodes := make([]*api.Node, 0, len(nodes))
	nodeInfos, err := framework.GetNodeInfoList(ctx, s.client)
	if err != nil {
		return nil, err
	}
	for _, nodeInfo := range nodeInfos {
		status := framework.NewStatus()
		for _, pl := range s.filterPlugins {
			status = pl.Filter(ctx, state, config, nodeInfo)
			if !status.IsSuccess() {
				status.SetFailedPlugin(pl.Name())
				break
			}
		}
		if status.IsSuccess() {
			feasibleNodes = append(feasibleNodes, nodeInfo.Node())
		}
	}
	return feasibleNodes, nil
}

func (s *Scheduler) RunScorePlugins(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nodes []*api.Node) (framework.NodeScoreList, *framework.Status) {
	s.logger.Info("scoring proxmox node")
	var scoresMap map[string](map[int]framework.NodeScore)
	nodeInfos, err := framework.GetNodeInfoList(ctx, s.client)
	if err != nil {
		status := framework.NewStatus()
		status.SetCode(1)
		return nil, status
	}
	for index, nodeInfo := range nodeInfos {
		for _, pl := range s.scorePlugins {
			score, status := pl.Score(ctx, state, config, nodeInfo)
			if !status.IsSuccess() {
				return nil, status
			}
			scoresMap[pl.Name()][index] = framework.NodeScore{
				Name:  nodeInfo.Node().Node,
				Score: score,
			}
		}
	}
	result := make(framework.NodeScoreList, 0, len(nodes))
	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i].Node, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}
	return result, nil
}

func selectHighestScoreNode(scoreList framework.NodeScoreList) (string, error) {
	if len(scoreList) == 0 {
		return "", fmt.Errorf("empty node score list")
	}
	selectedScore := framework.NodeScore{Score: -1}
	for _, nodescore := range scoreList {
		if selectedScore.Score < nodescore.Score {
			selectedScore = nodescore
		}
	}
	return selectedScore.Name, nil
}

func (s *Scheduler) RunVMIDPlugins(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nextid int, qemus []*api.VirtualMachine) (int, error) {

	// to do

	return nextid, nil
}
