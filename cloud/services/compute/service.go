package compute

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox"
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
		if IsNotFound(err) {
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
	vm, err := s.getVirtualMachineFromInstanceID(*instanceID)
	if err != nil && !proxmox.IsNotFound(err) {
		return nil, err
	} else if proxmox.IsNotFound(err) {
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

func (s *Service) getVirtualMachineFromInstanceID(instanceID string) (*proxmox.VirtualMachine, error) {
	nodeName, vmid := fetchVMIDFromInstanceID(instanceID)
	node, err := s.client.Node(nodeName)
	if err != nil {
		return nil, err
	}
	return node.VirtualMachine(vmid)
}

func (s *Service) CreateInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	node, err := s.GetRandomNode()
	if err != nil {
		return nil, err
	}

	// temp solution
	vmid := rand.Int()
	vm, err := node.CreateVirtualMachine(vmid)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func IsNotFound(err error) bool {
	return proxmox.IsNotFound(err)
}

func (s *Service) GetCluster() (*proxmox.Cluster, error) {
	return s.client.Cluster()
}

func (s *Service) GetNodes() ([]*proxmox.Node, error) {
	return s.client.Nodes()
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

func NewService(s scope.ProxmoxScope) *Service {
	return &Service{
		scope:  s,
		client: s.Client(),
	}
}
