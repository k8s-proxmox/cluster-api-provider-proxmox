package compute

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
	return nil
}

// wip
func (s *Service) createOrGetInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instance, err := s.GetCompute(ctx)
	if err != nil {
		return instance, err
	}
	return instance, err
}

// wip
func (s *Service) GetCompute(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instance := &proxmox.VirtualMachine{}
	node, err := s.client.Node("american")
	if err != nil {
		return nil, err
	}
	_ = log.FromContext(ctx)
	fmt.Printf("proxmox node : %v", node)
	return instance, err
}

func (s *Service) GetCluster() (*proxmox.Cluster, error) {
	return s.client.Cluster()
}

func (s *Service) GetNodes() ([]*proxmox.NodeStatus, error) {
	return s.client.Nodes()
}

// GetHostNode gets one node that will be used for vm host
// algorithm : random
func (s *Service) GetHostNode() (*proxmox.NodeStatus, error) {
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
