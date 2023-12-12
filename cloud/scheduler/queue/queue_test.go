package queue_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/queue"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/internal/fake"
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
		q.Add(&fake.QEMUSpec{})
	})
})

var _ = Describe("Get", Label("unit", "queue"), func() {
	var q *queue.SchedulingQueue

	BeforeEach(func() {
		q = queue.New()
	})

	Context("normal", func() {
		It("should run properly", func() {
			spec := &fake.QEMUSpec{}
			q.Add(spec)
			qemu, shutdown := q.Get()
			Expect(qemu).To(Equal(spec))
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
