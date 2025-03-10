package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/k8s-proxmox/proxmox-go/api"
	"github.com/k8s-proxmox/proxmox-go/proxmox"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/plugins"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/queue"
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
func NewManager(params SchedulerParams) (*Manager, error) {
	table := make(map[schedulerID]*Scheduler)
	config, err := plugins.GetPluginConfigFromFile(params.PluginConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %v", err)
	}
	params.pluginconfigs = config
	params.Logger.Info(fmt.Sprintf("load plugin config: %v", config))
	return &Manager{ctx: context.Background(), params: params, table: table}, nil
}

// return new/existing scheduler
func (m *Manager) GetOrCreateScheduler(client *proxmox.Service) *Scheduler {
	schedID, err := m.getSchedulerID(client)
	if err != nil {
		// create new scheduler without registering
		// to not make it zombie scheduler set timeout to context
		sched := m.NewScheduler(client, WithTimeout(1*time.Minute))
		return sched
	}
	sched, ok := m.table[*schedID]
	if !ok {
		// create and register new scheduler
		m.params.Logger.V(4).Info("registering new scheduler")
		sched := m.NewScheduler(client)
		sched.logger = sched.logger.WithValues("schedulerID", &schedID)
		m.table[*schedID] = sched
		return sched
	}
	sched.logger.V(4).Info("using existing scheduler")
	return sched
}

// return new scheduler.
// usually better to use GetOrCreateScheduler instead.
func (m *Manager) NewScheduler(client *proxmox.Service, opts ...SchedulerOption) *Scheduler {
	ctx, cancel := context.WithCancel(m.ctx)
	sched := &Scheduler{
		client:          client,
		schedulingQueue: queue.New(),

		registry: plugins.NewRegistry(m.params.PluginConfigs()),

		resultMap: make(map[string]chan *framework.CycleState),
		logger:    m.params.Logger.WithValues("Name", "qemu-scheduler"),

		ctx:    ctx,
		cancel: cancel,
	}

	for _, fn := range opts {
		fn(sched)
	}

	return sched
}

type SchedulerOption func(s *Scheduler)
type CancelFunc func()

func (s *Scheduler) WithCancel() (*Scheduler, CancelFunc) {
	return s, s.Stop
}

// set timeout to scheduler
func WithTimeout(timeout time.Duration) SchedulerOption {
	return func(s *Scheduler) {
		_, cancel := s.WithCancel()
		go time.AfterFunc(timeout, cancel)
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

	registry plugins.PluginRegistry

	// to do : cache

	// map[qemu name]chan *framework.CycleState
	resultMap map[string]chan *framework.CycleState
	logger    logr.Logger

	// scheduler status
	running bool

	// scheduler runs until this context done
	ctx context.Context

	// to stop itself
	cancel context.CancelFunc
}

type SchedulerParams struct {
	Logger logr.Logger

	// file path for pluginConfig
	PluginConfigFile string
	pluginconfigs    plugins.PluginConfigs
}

func (p *SchedulerParams) PluginConfigs() plugins.PluginConfigs {
	return p.pluginconfigs
}

type schedulerID struct {
	IPAddress   string
	Fingreprint string
}

// run scheduler
// and ensure only one process is running
func (s *Scheduler) Run() {
	if s.IsRunning() {
		s.logger.Info("this scheduler is already running")
		return
	}
	defer func() { s.running = false }()
	s.running = true
	s.logger.Info("Start Running Scheduler")
	wait.UntilWithContext(s.ctx, s.ScheduleOne, 0)
	s.logger.Info("Stop Running Scheduler")
}

func (s *Scheduler) IsRunning() bool {
	return s.running
}

// run scheduelr in parallel
func (s *Scheduler) RunAsync() {
	go s.Run()
}

// stop scheduler
func (s *Scheduler) Stop() {
	defer s.cancel()
	s.schedulingQueue.ShutDown()
}

// retrieve one qemuSpec from queue and try to create
// new qemu according to the qemuSpec
func (s *Scheduler) ScheduleOne(ctx context.Context) {
	s.logger.Info("getting next qemu from scheduling queue")
	qemu, shutdown := s.schedulingQueue.Get()
	if shutdown {
		return
	}
	config := qemu.Config()
	qemuCtx := qemu.Context()
	s.logger.Info("scheduling qemu")

	state := framework.NewCycleState()
	s.resultMap[config.Name] = make(chan *framework.CycleState, 1)
	defer func() { s.resultMap[config.Name] <- &state }()

	// select node to run qemu
	node, err := s.SelectNode(qemuCtx, *config)
	if err != nil {
		state.UpdateState(true, err, framework.SchedulerResult{})
		return
	}

	// select vmid to be assigned to qemu
	// to do: do this in parallel with SelectNode
	vmid, err := s.SelectVMID(qemuCtx, *config)
	if err != nil {
		state.UpdateState(true, err, framework.SchedulerResult{})
		return
	}

	// select vm storage to be used for vm image
	// must be done after node selection as some storages may not be available on some nodes
	storage, err := s.SelectStorage(qemuCtx, *config, node)
	if err != nil {
		state.UpdateState(true, err, framework.SchedulerResult{})
		return
	}

	result := framework.NewSchedulerResult(vmid, node, storage)
	state.UpdateState(true, nil, result)
}

// wait until CycleState is put into channel and then return it
func (s *Scheduler) WaitStatus(ctx context.Context, config *api.VirtualMachineCreateOptions) (framework.CycleState, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	var done chan *framework.CycleState
	ok := false
	for !ok {
		done, ok = s.resultMap[config.Name]
		if !ok {
			time.Sleep(100 * time.Millisecond)
		}
	}
	select {
	case state := <-done:
		delete(s.resultMap, config.Name)
		return *state, nil
	case <-ctx.Done():
		err := fmt.Errorf("exceed timeout deadline. schedulingQueue might be shutdowned")
		s.logger.Error(err, fmt.Sprintf("schedulingQueue: %v", *s.schedulingQueue))
		return framework.CycleState{}, err
	}
}

// create new qemu with given spec and context
func (s *Scheduler) CreateQEMU(ctx context.Context, config *api.VirtualMachineCreateOptions) (framework.SchedulerResult, error) {
	log := s.logger.WithValues("qemu", config.Name)
	log.Info("adding qemu to scheduler queue")
	// add qemu spec into the queue
	s.schedulingQueue.Add(ctx, config)

	// wait until the scheduller finishes its job
	var err error
	status, err := s.WaitStatus(ctx, config)
	if err != nil {
		return status.Result(), err
	}
	if status.Error() != nil {
		log.Error(status.Error(), fmt.Sprintf("failed to create qemu: %v", status.Messages()))
		return status.Result(), status.Error()
	}
	log.Info(fmt.Sprintf("%v", status.Messages()))
	return status.Result(), nil
}

func (s *Scheduler) SelectNode(ctx context.Context, config api.VirtualMachineCreateOptions) (string, error) {
	s.logger.Info("finding proxmox node matching qemu")
	nodes, err := s.client.GetNodes(ctx)
	if err != nil {
		return "", err
	}

	state := framework.NewCycleState()

	// filter
	nodelist, _ := s.RunFilterPlugins(ctx, &state, config, nodes)
	if len(nodelist) == 0 {
		return "", ErrNoNodesAvailable
	}
	if len(nodelist) == 1 {
		return nodelist[0].Node, nil
	}

	// score
	scorelist, status := s.RunScorePlugins(ctx, &state, config, nodelist)
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
	usedID, err := usedIDMap(ctx, s.client)
	if err != nil {
		return 0, err
	}
	return s.RunVMIDPlugins(ctx, nil, config, nextid, *usedID)
}

func (s *Scheduler) SelectStorage(ctx context.Context, config api.VirtualMachineCreateOptions, nodeName string) (string, error) {
	log := s.logger.WithValues("qemu", config.Name).WithValues("node", nodeName)
	log.Info("finding proxmox storage to be used for qemu")
	if config.Storage != "" {
		// to do: raise error if storage is not available on the node
		return config.Storage, nil
	}

	node, err := s.client.Node(ctx, nodeName)
	if err != nil {
		log.Error(err, "failed to get node")
		return "", err
	}
	storages, err := node.GetStorages(ctx)
	if err != nil {
		log.Error(err, "failed to get storages")
		return "", err
	}

	// current logic is just selecting the first storage
	// that is active and supports "images" type of content
	for _, storage := range storages {
		if strings.Contains(storage.Content, "images") && storage.Active == 1 {
			return storage.Storage, nil
		}
	}

	return "", fmt.Errorf("no storage available for VM image on node %s", nodeName)
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
		for _, pl := range s.registry.FilterPlugins() {
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

func (s *Scheduler) RunScorePlugins(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nodes []*api.Node) (map[string]framework.NodeScore, *framework.Status) {
	s.logger.Info("scoring proxmox node")
	status := framework.NewStatus()
	scoresMap := make(map[string](map[string]framework.NodeScore))
	for _, pl := range s.registry.ScorePlugins() {
		scoresMap[pl.Name()] = make(map[string]framework.NodeScore)
	}
	nodeInfos, err := framework.GetNodeInfoList(ctx, s.client)
	if err != nil {
		status.SetCode(1)
		s.logger.Error(err, "failed to get node info list")
		return nil, status
	}
	for _, nodeInfo := range nodeInfos {
		for _, pl := range s.registry.ScorePlugins() {
			score, status := pl.Score(ctx, state, config, nodeInfo)
			if !status.IsSuccess() {
				status.SetCode(1)
				s.logger.Error(status.Error(), fmt.Sprintf("failed to score node %s", nodeInfo.Node().Node))
				return nil, status
			}
			scoresMap[pl.Name()][nodeInfo.Node().Node] = framework.NodeScore{
				Name:  nodeInfo.Node().Node,
				Score: score,
			}
		}
	}
	result := make(map[string]framework.NodeScore)
	for _, node := range nodes {
		result[node.Node] = framework.NodeScore{Name: node.Node, Score: 0}
		for plugin := range scoresMap {
			r := result[node.Node]
			r.Score += scoresMap[plugin][node.Node].Score
		}
	}
	return result, status
}

func selectHighestScoreNode(scoreList map[string]framework.NodeScore) (string, error) {
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

func (s *Scheduler) RunVMIDPlugins(ctx context.Context, state *framework.CycleState, config api.VirtualMachineCreateOptions, nextid int, usedID map[int]bool) (int, error) {
	for _, pl := range s.registry.VMIDPlugins() {
		key := pl.PluginKey()
		value := ctx.Value(key)
		if value != nil {
			s.logger.WithValues("vmid plugin", pl.Name()).Info("selecting vmid")
			return pl.Select(ctx, state, config, nextid, usedID)
		}
	}
	s.logger.Info("no vmid key found. using nextid")
	return nextid, nil
}

// return map[vmid]bool
func usedIDMap(ctx context.Context, client *proxmox.Service) (*map[int]bool, error) {
	vms, err := client.VirtualMachines(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[int]bool)
	for _, vm := range vms {
		result[vm.VMID] = true
	}
	return &result, nil
}
