package instance

import (
	"context"

	"github.com/sp-yduck/proxmox-go/proxmox"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler"
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
		scheduler: s.GetScheduler().WithClient(s.CloudClient()),
	}
}

func (s *Service) vncClient(nodeName string) (*proxmox.VNCWebSocketClient, error) {
	client, err := s.client.NewNodeVNCWebSocketConnection(context.TODO(), nodeName)
	if err != nil {
		return nil, err
	}
	return client, nil
}
