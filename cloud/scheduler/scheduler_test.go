package scheduler_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sp-yduck/proxmox-go/api"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
)

var _ = Describe("NewManager", Label("unit", "scheduler"), func() {
	It("should not error", func() {
		params := scheduler.SchedulerParams{}
		manager := scheduler.NewManager(params)
		Expect(manager).NotTo(BeNil())
	})
})

var _ = Describe("NewScheduler", Label("unit", "scheduler"), func() {
	manager := scheduler.NewManager(scheduler.SchedulerParams{})

	It("should not error", func() {
		sched := manager.NewScheduler(proxmoxSvc)
		Expect(sched).NotTo(BeNil())
	})
})

var _ = Describe("GetOrCreateScheduler", Label("integration", "scheduler"), func() {
	manager := scheduler.NewManager(scheduler.SchedulerParams{})

	It("should not error", func() {
		sched := manager.GetOrCreateScheduler(proxmoxSvc)
		Expect(sched).NotTo(BeNil())
	})
})

var _ = Describe("Run (RunAsync) / IsRunning / Stop", Label("unit", "scheduler"), func() {
	manager := scheduler.NewManager(scheduler.SchedulerParams{zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))})

	Context("with minimal scheduler", func() {
		It("should not error", func() {
			sched := manager.NewScheduler(proxmoxSvc)
			sched.RunAsync()
			time.Sleep(1 * time.Second)
			Expect(sched.IsRunning()).To(BeTrue())
			sched.Stop()
			time.Sleep(1 * time.Second)
			Expect(sched.IsRunning()).To(BeFalse())
		})
	})
})

var _ = Describe("WithTimeout", Label("integration", "scheduler"), func() {
	manager := scheduler.NewManager(scheduler.SchedulerParams{})

	It("should not error", func() {
		sched := manager.NewScheduler(proxmoxSvc, scheduler.WithTimeout(2*time.Second))
		Expect(sched).NotTo(BeNil())
		sched.RunAsync()
		time.Sleep(1 * time.Second)
		Expect(sched.IsRunning()).To(BeTrue())
		time.Sleep(2 * time.Second)
		Expect(sched.IsRunning()).To(BeFalse())
	})
})

var _ = Describe("CreateQEMU", Label("integration"), func() {
	manager := scheduler.NewManager(scheduler.SchedulerParams{zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))})
	var result framework.SchedulerResult

	AfterEach(func() {
		vm, err := proxmoxSvc.VirtualMachine(context.Background(), result.VMID())
		if err == nil {
			err := vm.Delete(context.Background())
			Expect(err).NotTo(HaveOccurred())
		}
	})

	Context("with minimal scheduler", func() {
		It("should not error", func() {
			sched := manager.NewScheduler(proxmoxSvc)
			sched.RunAsync()
			var err error
			result, err = sched.CreateQEMU(context.Background(), &api.VirtualMachineCreateOptions{
				Name: "qemu-scheduler-test-createqemu",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Node()).To(Equal(result.Instance().Node))
			Expect(result.VMID()).To(Equal(result.Instance().VM.VMID))
		})
	})
})
