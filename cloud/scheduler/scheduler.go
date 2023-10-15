package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// ErrNoNodesAvailable is used to describe the error that no nodes available to schedule qemus.
	ErrNoNodesAvailable = fmt.Errorf("no nodes available to schedule qemus")

	// ErrNoVMIDAvailable is used to describe the error that no vmid available to schedule qemus.
	ErrNoVMIDAvailable = fmt.Errorf("no vmid available to schedule qemus")
)

type Scheduler struct {
	k8sClient client.Client
	client    *proxmox.Service

	// logger *must* be initialized when creating a Scheduler,
	// otherwise logging functions will access a nil sink and
	// panic.
	logger klog.Logger
}

type ScheduleResult struct {
}

func New(k8sClient client.Client) *Scheduler {
	ctx := context.WithValue(context.Background(), "component", "qemu-scheduler")
	logger := klog.FromContext(ctx)
	return &Scheduler{k8sClient: k8sClient, logger: logger}
}

func (s *Scheduler) WithClient(client *proxmox.Service) *Scheduler {
	s.client = client
	return s
}

// just poc codes
// return randomly chosen node
func (s *Scheduler) GetNode(ctx context.Context) (*api.Node, error) {
	if s.client == nil {
		return nil, fmt.Errorf("proxmox client is empty")
	}
	nodes, err := s.client.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) <= 0 {
		return nil, ErrNoNodesAvailable
	}
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	return nodes[r.Intn(len(nodes))], nil
}

// just poc codes
// return nextID fetched from Proxmox rest API nextID endpoint
func (s *Scheduler) GetID(ctx context.Context) (int, error) {
	return s.client.RESTClient().GetNextID(ctx)
}

func (s *Scheduler) ScheduleQEMU() (ScheduleResult, error) {
	return ScheduleResult{}, nil
}
