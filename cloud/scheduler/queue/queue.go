package queue

import (
	"context"
	"sync"

	"github.com/sp-yduck/proxmox-go/api"
)

type SchedulingQueue struct {
	activeQ      []*qemuSpec
	lock         *sync.Cond
	shuttingDown bool
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

	if s.shuttingDown {
		return
	}

	s.activeQ = append(s.activeQ, &qemuSpec{ctx: ctx, config: config})
	s.lock.Signal()
}

// return length of active queue
// func (s *SchedulingQueue) Len() int {
// 	s.lock.L.Lock()
// 	defer s.lock.L.Unlock()
// 	return len(s.activeQ)
// }

// return nex qemuSpec
func (s *SchedulingQueue) Get() (spec *qemuSpec, shutdown bool) {
	s.lock.L.Lock()
	defer s.lock.L.Unlock()
	for len(s.activeQ) == 0 && !s.shuttingDown {
		s.lock.Wait()
	}
	if len(s.activeQ) == 0 {
		return nil, true
	}

	spec = s.activeQ[0]
	// The underlying array still exists and reference this object,
	// so the object will not be garbage collected.
	s.activeQ[0] = nil
	s.activeQ = s.activeQ[1:]
	return spec, false
}

// shut down the queue
func (s *SchedulingQueue) ShutDown() {
	s.lock.L.Lock()
	defer s.lock.L.Unlock()
	s.shuttingDown = true
	s.lock.Broadcast()
}

func (s *qemuSpec) Config() *api.VirtualMachineCreateOptions {
	return s.config
}

func (s *qemuSpec) Context() context.Context {
	return s.ctx
}
