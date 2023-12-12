package queue

import (
	"context"
	"sync"

	"github.com/k8s-proxmox/proxmox-go/api"
)

type SchedulingQueue struct {
	activeQ      []QEMUSpec
	lock         *sync.Cond
	shuttingDown bool
}

func New() *SchedulingQueue {
	return &SchedulingQueue{
		activeQ: []QEMUSpec{},
		lock:    sync.NewCond(&sync.Mutex{}),
	}
}

// qemu create option and context.
// each scheduling plugins retrieves values from this context
type QEMUSpec interface {
	Name() string
	Context() context.Context
	CloneSpec() *api.VirtualMachineCloneOption
	Config() *api.VirtualMachineConfig
	CreateSpec() *api.VirtualMachineCreateOptions
}

// add new qemuSpec to queue
func (s *SchedulingQueue) Add(qemu QEMUSpec) {
	s.lock.L.Lock()
	defer s.lock.L.Unlock()

	if s.shuttingDown {
		return
	}

	s.activeQ = append(s.activeQ, qemu)
	s.lock.Signal()
}

// return length of active queue
// func (s *SchedulingQueue) Len() int {
// 	s.lock.L.Lock()
// 	defer s.lock.L.Unlock()
// 	return len(s.activeQ)
// }

// return next QEMUSpec
func (s *SchedulingQueue) Get() (spec QEMUSpec, shutdown bool) {
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
