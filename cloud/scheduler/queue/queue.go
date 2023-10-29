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

// qemu create option and context.
// each scheduling plugins retrieves values from this context
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
// to do : break if context is done
func (s *SchedulingQueue) NextQEMU(ctx context.Context) *qemuSpec {
	// wait
	s.lock.L.Lock()
	defer s.lock.L.Unlock()
	for len(s.activeQ) == 0 {
		s.lock.Wait()
	}
	spec := s.activeQ[0]
	s.activeQ = s.activeQ[1:]
	return spec
}

func (s *qemuSpec) Config() *api.VirtualMachineCreateOptions {
	return s.config
}

func (s *qemuSpec) Context() context.Context {
	return s.ctx
}
