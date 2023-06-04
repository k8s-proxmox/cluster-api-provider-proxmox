package compute

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"

	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling instance resources")
	instance, err := s.createOrGetInstance(ctx)
	if err != nil {
		log.Error(err, "failed to create/get instance")
		return err
	}
	log.Info(fmt.Sprintf("instance : %v", instance))

	s.scope.SetProviderID(instance)
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))

	// ensure instance is running
	switch instance.Status {
	case vm.ProcessStatusRunning:
		return nil
	case vm.ProcessStatusStopped:
		if err := instance.Start(vm.StartOption{}); err != nil {
			return err
		}
	case vm.ProcessStatusPaused:
		if err := instance.Resume(vm.ResumeOption{}); err != nil {
			return err
		}
	default:
		return errors.Errorf("unexpected status : %s", instance.Status)
	}
	return nil
}

func (s *Service) createOrGetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)
	if s.scope.GetInstanceID() == nil {
		log.Info("ProxmoxMachine doesn't have instanceID. instance will be created")
		return s.CreateInstance(ctx)
	}
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if IsNotFound(err) {
			log.Info("instance wasn't found. new instance will be created")
			return s.CreateInstance(ctx)
		}
		log.Error(err, "failed to get instance")
		return nil, err
	}
	return instance, nil
}

func (s *Service) GetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)
	instanceID := s.scope.GetInstanceID()
	vm, err := s.getInstanceFromInstanceID(*instanceID)
	if err != nil && !api.IsNotFound(err) {
		log.Error(err, "failed to get instance from instance ID")
		return nil, err
	} else if api.IsNotFound(err) {
		log.Info("instance wasn't found")
		return nil, api.ErrNotFound
	}
	return vm, nil
}

func (s *Service) getInstanceFromInstanceID(instanceID string) (*vm.VirtualMachine, error) {
	vmid, err := strconv.Atoi(instanceID)
	if err != nil {
		return nil, err
	}
	nodes, err := s.client.Nodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("proxmox nodes not found")
	}
	for _, node := range nodes {
		vm, err := node.VirtualMachine(vmid)
		if err != nil {
			continue
		}
		return vm, nil
	}
	return nil, api.ErrNotFound
}

func (s *Service) CreateInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// temp solution
	node, err := s.GetRandomNode()
	if err != nil {
		log.Error(err, "failed to get random node")
		return nil, err
	}

	vmid, err := s.GetNextID()
	if err != nil {
		log.Error(err, "failed to get availabel vmid")
		return nil, err
	}

	vmoption := vm.VirtualMachineCreateOptions{}
	vm, err := node.CreateVirtualMachine(vmid, vmoption)
	if err != nil {
		log.Error(err, "failed to create virtual machine")
		return nil, err
	}
	return vm, nil
}

func IsNotFound(err error) bool {
	return api.IsNotFound(err)
}

func (s *Service) GetNextID() (int, error) {
	return s.client.NextID()
}

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

func (s *Service) Delete(ctx context.Context) error {
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		return nil
	}
	return instance.Delete()
}
