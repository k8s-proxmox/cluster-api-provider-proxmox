package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/queue"
	"github.com/k8s-proxmox/proxmox-go/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueue(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue Suite")
}

var _ = Describe("New", Label("unit", "queue"), func() {
	It("shoud not error", func() {
		q := queue.New()
		Expect(q).ToNot(BeNil())
	})
})

var _ = Describe("Add", Label("unit", "queue"), func() {
	q := queue.New()

	It("should not error", func() {
		q.Add(context.Background(), &api.VirtualMachineCreateOptions{Name: "foo"})
	})
})

var _ = Describe("Get", Label("unit", "queue"), func() {
	var q *queue.SchedulingQueue

	BeforeEach(func() {
		q = queue.New()
	})

	Context("normal", func() {
		It("should run properly", func() {
			c := &api.VirtualMachineCreateOptions{Name: "foo"}
			q.Add(context.Background(), c)
			qemu, shutdown := q.Get()
			Expect(qemu.Config()).To(Equal(c))
			Expect(shutdown).To(BeFalse())
		})
	})

	Context("shutdown empty queue after 1 sec", func() {
		It("should get nil", func() {
			go func() {
				time.Sleep(1 * time.Second)
				q.ShutDown()
			}()
			qemu, shutdown := q.Get()
			Expect(qemu).To(BeNil())
			Expect(shutdown).To(BeTrue())
		})
	})
})
