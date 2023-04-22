package vm

import (
	"context"
	"fmt"

	"github.com/luthermonson/go-proxmox"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

type Service struct {
	scope  scope.ProxmoxScope
	client proxmox.Client
}

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

func (s *Service) createOrGetInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instance, err := s.GetCompute(ctx)
	if err != nil {
		return instance, err
	}
	return instance, err
}

func (s *Service) GetCompute(ctx context.Context) (*proxmox.VirtualMachine, error) {
	instance := &proxmox.VirtualMachine{}
	node, err := s.client.Node("american")
	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("proxmox node : %v", node))
	return instance, err
}

func (s *Service) Delete(ctx context.Context) error {
	return nil
}

func NewService(s scope.ProxmoxScope) *Service {

	return &Service{
		scope:  s,
		client: s.Client(),
	}
}
