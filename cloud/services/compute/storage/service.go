package storage

import (
	"github.com/sp-yduck/proxmox-go/proxmox"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
)

type Scope interface {
	cloud.Cluster
}

type Service struct {
	scope  Scope
	client proxmox.Service
}

func NewService(s Scope) *Service {
	return &Service{
		scope:  s,
		client: *s.CloudClient(),
	}
}
