package fake

import (
	"context"

	"github.com/k8s-proxmox/proxmox-go/api"
)

// implementation for queue.QEMUSpec
type QEMUSpec struct {
	ctx        context.Context
	name       string
	node       string
	vmid       *int
	createSpec *api.VirtualMachineCreateOptions
	cloneSpec  *api.VirtualMachineCloneOption
	config     *api.VirtualMachineConfig
}

func (s *QEMUSpec) Name() string {
	return s.name
}

func (s *QEMUSpec) CreateSpec() *api.VirtualMachineCreateOptions {
	return s.createSpec
}

func (s *QEMUSpec) Context() context.Context {
	return s.ctx
}

func (s *QEMUSpec) CloneSpec() *api.VirtualMachineCloneOption {
	return s.cloneSpec
}

func (s *QEMUSpec) Config() *api.VirtualMachineConfig {
	return s.config
}
