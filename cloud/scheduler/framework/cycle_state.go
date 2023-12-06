package framework

import (
	"github.com/k8s-proxmox/proxmox-go/api"
	"github.com/k8s-proxmox/proxmox-go/proxmox"
)

type CycleState struct {
	completed bool
	err       error
	messages  map[string]string
	result    SchedulerResult
}

type SchedulerResult struct {
	vmid     int
	node     string
	instance *proxmox.VirtualMachine
}

func NewCycleState() CycleState {
	return CycleState{completed: false, err: nil, messages: map[string]string{}}
}

func (c *CycleState) SetComplete() {
	c.completed = true
}

func (c *CycleState) IsCompleted() bool {
	return c.completed
}

func (c *CycleState) SetError(err error) {
	c.err = err
}

func (c *CycleState) Error() error {
	return c.err
}

func (c *CycleState) SetMessage(pluginName, message string) {
	c.messages[pluginName] = message
}

func (c *CycleState) Messages() map[string]string {
	return c.messages
}

func (c *CycleState) QEMU() *api.VirtualMachine {
	return c.result.instance.VM
}

func (c *CycleState) UpdateState(completed bool, err error, result SchedulerResult) {
	c.completed = completed
	c.err = err
	c.result = result
}

func NewSchedulerResult(vmid int, node string, instance *proxmox.VirtualMachine) SchedulerResult {
	return SchedulerResult{vmid: vmid, node: node, instance: instance}
}

func (c *CycleState) Result() SchedulerResult {
	return c.result
}

func (r *SchedulerResult) Node() string {
	return r.node
}

func (r *SchedulerResult) VMID() int {
	return r.vmid
}

func (r *SchedulerResult) Instance() *proxmox.VirtualMachine {
	return r.instance
}
