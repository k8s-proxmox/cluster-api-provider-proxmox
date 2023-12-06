package storage

import (
	"github.com/k8s-proxmox/proxmox-go/proxmox"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud"
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
