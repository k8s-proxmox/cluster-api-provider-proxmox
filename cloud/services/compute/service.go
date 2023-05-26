package compute

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

type Service struct {
	scope  scope.ProxmoxScope
	client proxmox.Client
}

// wip
func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling instance resources")
	instance, err := s.createOrGetInstance(ctx)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("instance : %v", instance))

	s.scope.SetProviderID()
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))

	return nil
}

// wip
func (s *Service) createOrGetInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if IsNotFoundError(err) {
			instance, err = s.CreateInstance(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return instance, err
}

func (s *Service) GetInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instanceID := s.scope.GetInstanceID()
	nodes, err := s.GetNodes()
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		vm, err := getInstanceFromInstanceID(*node, *instanceID)
		if err == nil && vm != nil {
			return vm, nil
		}
	}
	return nil, errors.New("no resource found")
}

func (s *Service) CreateInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	node, err := s.GetRandomNode()
	if err != nil {
		return nil, err
	}

	// temp solution
	vmid := rand.Int()
	option := proxmox.VirtualMachineOption{
		Name: *s.scope.GetInstanceID(),
	}

	_, err = node.NewVirtualMachine(vmid, option)
	if err != nil {
		return nil, err
	}

	vm, err := getInstanceFromVMID(*node, vmid)
	if err != nil {

		return nil, err
	}
	return vm, nil
}

func IsNotFoundError(err error) bool {
	return err.Error() == "no resource found"
}

func (s *Service) GetCluster() (*proxmox.Cluster, error) {
	return s.client.Cluster()
}

func (s *Service) GetNodes() ([]*proxmox.Node, error) {
	ns, err := s.client.Nodes()
	if err != nil {
		return nil, err
	}
	var nodes []*proxmox.Node
	for _, n := range ns {
		node, err := s.client.Node(n.Node)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func getInstanceFromVMID(node proxmox.Node, vmid int) (*proxmox.VirtualMachine, error) {
	return node.VirtualMachine(vmid)
}

func getInstanceFromInstanceID(node proxmox.Node, instanceID string) (*proxmox.VirtualMachine, error) {
	vms, err := node.VirtualMachines()
	if err != nil {
		return nil, err
	}
	for _, vm := range vms {
		if vm.Name == instanceID {
			return vm, nil
		}
	}
	return nil, errors.New("no resource found")
}

// GetRandomNode returns a node chosen randomly
func (s *Service) GetRandomNode() (*proxmox.Node, error) {
	nodes, err := s.GetNodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) <= 0 {
		return nil, errors.Errorf("no nodes found")
	}
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	node := nodes[r.Intn(len(nodes))]
	return node, nil
}

// wip
func (s *Service) Delete(ctx context.Context) error {
	return nil
}

func NewService(s scope.ProxmoxScope) *Service {
	return &Service{
		scope:  s,
		client: s.Client(),
	}
}
