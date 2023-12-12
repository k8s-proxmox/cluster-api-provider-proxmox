package scheduler_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/internal/fake"
)

var _ = Describe("NewManager", Label("unit", "scheduler"), func() {
	Context("with empty params", func() {
		It("should not error", func() {
			params := scheduler.SchedulerParams{}
			manager, err := scheduler.NewManager(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager).NotTo(BeNil())
		})
	})
	Context("with only logger", func() {
		It("should not error", func() {
			params := scheduler.SchedulerParams{Logger: zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))}
			manager, err := scheduler.NewManager(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager).NotTo(BeNil())
		})
	})
	Context("with plugin-config", func() {
		It("should not error", func() {
			params := scheduler.SchedulerParams{PluginConfigFile: ""}
			manager, err := scheduler.NewManager(params)
			Expect(err).NotTo(HaveOccurred())
			Expect(manager).NotTo(BeNil())
		})
	})
})

var _ = Describe("NewScheduler", Label("unit", "scheduler"), func() {
	manager, err := scheduler.NewManager(scheduler.SchedulerParams{})
	Expect(err).NotTo(HaveOccurred())

	It("should not error", func() {
		sched := manager.NewScheduler(proxmoxSvc)
		Expect(sched).NotTo(BeNil())
	})
})

var _ = Describe("GetOrCreateScheduler", Label("integration", "scheduler"), func() {
	manager, err := scheduler.NewManager(scheduler.SchedulerParams{})
	Expect(err).NotTo(HaveOccurred())

	It("should not error", func() {
		sched := manager.GetOrCreateScheduler(proxmoxSvc)
		Expect(sched).NotTo(BeNil())
	})
})

var _ = Describe("Run (RunAsync) / IsRunning / Stop", Label("unit", "scheduler"), func() {
	manager, err := scheduler.NewManager(scheduler.SchedulerParams{Logger: zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))})
	Expect(err).NotTo(HaveOccurred())

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
	manager, err := scheduler.NewManager(scheduler.SchedulerParams{})
	Expect(err).NotTo(HaveOccurred())

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
	manager, err := scheduler.NewManager(scheduler.SchedulerParams{Logger: zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))})
	Expect(err).NotTo(HaveOccurred())
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
			result, err = sched.CreateQEMU(&fake.QEMUSpec{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Node()).To(Equal(result.Instance().Node))
			Expect(result.VMID()).To(Equal(result.Instance().VM.VMID))
		})
	})
})
