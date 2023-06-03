package compute

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service"
	"github.com/sp-yduck/proxmox/pkg/service/node"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
)

type Scope interface {
	cloud.Machine
}

type Service struct {
	scope  Scope
	client service.Service
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

	s.scope.SetProviderID(instance)
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))
	return nil
}

// wip
func (s *Service) createOrGetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	if s.scope.GetInstanceID() == nil {
		return s.CreateInstance(ctx)
	}
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if IsNotFound(err) {
			return s.CreateInstance(ctx)
		}
		return nil, err
	}
	return instance, nil
}

func (s *Service) GetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	instanceID := s.scope.GetInstanceID()
	vm, err := s.getVirtualMachineFromInstanceID(*instanceID)
	if err != nil && !api.IsNotFound(err) {
		return nil, err
	} else if api.IsNotFound(err) {
		return nil, errors.New("no resource found")
	}
	return vm, nil
}

func fetchVMIDFromInstanceID(instanceID string) (string, int) {
	s := strings.Split(instanceID, "/")
	nodeName := s[0]
	vmid, _ := strconv.Atoi(s[1])
	return nodeName, vmid
}

func (s *Service) getVirtualMachineFromInstanceID(instanceID string) (*vm.VirtualMachine, error) {
	nodeName, vmid := fetchVMIDFromInstanceID(instanceID)
	node, err := s.client.Node(nodeName)
	if err != nil {
		return nil, err
	}
	return node.VirtualMachine(vmid)
}

func (s *Service) CreateInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	node, err := s.GetRandomNode()
	if err != nil {
		return nil, err
	}

	// temp solution
	vmid := rand.Intn(99999)
	klog.Infof("vmid : %d", vmid)
	vmoption := vm.VirtualMachineCreateOptions{}
	vm, err := node.CreateVirtualMachine(vmid, vmoption)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func IsNotFound(err error) bool {
	return api.IsNotFound(err)
}

// func (s *Service) GetCluster() (*service.Cluster, error) {
// 	return s.client.Cluster()
// }

func (s *Service) GetNodes() ([]*node.Node, error) {
	return s.client.Nodes()
}

// GetRandomNode returns a node chosen randomly
func (s *Service) GetRandomNode() (*node.Node, error) {
	nodes, err := s.GetNodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) <= 0 {
		return nil, errors.Errorf("no nodes found")
	}
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	return nodes[r.Intn(len(nodes))], nil
}

// wip
func (s *Service) Delete(ctx context.Context) error {
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		return nil
	}
	_, err = instance.Delete()
	if err != nil {
		return err
	}
	return nil
}

func NewService(s Scope) *Service {
	return &Service{
		scope:  s,
		client: *s.CloudClient(),
	}
}
