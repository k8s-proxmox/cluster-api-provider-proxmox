package instance

import (
	"context"

	"github.com/k8s-proxmox/proxmox-go/proxmox"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler"
)

type Scope interface {
	cloud.Machine
}

type Service struct {
	scope     Scope
	client    proxmox.Service
	scheduler *scheduler.Scheduler
}

func NewService(s Scope) *Service {
	return &Service{
		scope:     s,
		client:    *s.CloudClient(),
		scheduler: s.GetScheduler(s.CloudClient()),
	}
}

func (s *Service) vncClient(nodeName string) (*proxmox.VNCWebSocketClient, error) {
	client, err := s.client.NewNodeVNCWebSocketConnection(context.TODO(), nodeName)
	if err != nil {
		return nil, err
	}
	return client, nil
}
