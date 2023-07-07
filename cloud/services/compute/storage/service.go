package storage

import (
	"github.com/sp-yduck/proxmox/pkg/service"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
)

type Scope interface {
	cloud.Cluster
}

type Service struct {
	scope  Scope
	client service.Service
}

func NewService(s Scope) *Service {
	return &Service{
		scope:  s,
		client: *s.CloudClient(),
	}
}
