package queue

import (
	"context"
	"sync"

	"github.com/sp-yduck/proxmox-go/api"
)

type SchedulingQueue struct {
	activeQ []*qemuSpec
	lock    *sync.Cond
}

func New() *SchedulingQueue {
	return &SchedulingQueue{
		activeQ: []*qemuSpec{},
		lock:    sync.NewCond(&sync.Mutex{}),
	}
}

// qemu create option with context
type qemuSpec struct {
	ctx    context.Context
	config *api.VirtualMachineCreateOptions
}

// add new qemuSpec to queue
func (s *SchedulingQueue) Add(ctx context.Context, config *api.VirtualMachineCreateOptions) {
	s.lock.L.Lock()
	defer s.lock.L.Unlock()
	s.activeQ = append(s.activeQ, &qemuSpec{ctx: ctx, config: config})
	s.lock.Signal()
}

// return next qemuSpec
func (s *SchedulingQueue) NextQEMU() *qemuSpec {
	// wait
	s.lock.L.Lock()
	for len(s.activeQ) == 0 {
		s.lock.Wait()
	}
	spec := s.activeQ[0]
	s.activeQ = s.activeQ[1:]
	s.lock.L.Unlock()
	return spec
}

func (s *qemuSpec) Config() *api.VirtualMachineCreateOptions {
	return s.config
}

func (s *qemuSpec) Context() context.Context {
	return s.ctx
}
