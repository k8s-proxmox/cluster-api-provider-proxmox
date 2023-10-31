package idrange_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/idrange"
)

func TestIDRange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "idrange plugin")
}

var _ = Describe("findVMIDRange", Label("unit", "plugins"), func() {
	ctx := context.Background()

	Context("no 'vmid.qemu-scheduler/range' key in the context", func() {
		It("should error", func() {
			start, end, err := idrange.FindVMIDRange(ctx)
			Expect(start).To(Equal(0))
			Expect(end).To(Equal(0))
			Expect(err.Error()).To(Equal("no vmid range is specified"))
		})
	})

	Context("specify invalid range (start)", func() {
		It("should error", func() {
			c := context.WithValue(ctx, framework.CtxKey(idrange.VMIDRangeKey), "a-10")
			start, end, err := idrange.FindVMIDRange(c)
			Expect(start).To(Equal(0))
			Expect(end).To(Equal(0))
			Expect(err.Error()).To(ContainSubstring("invalid range is specified"))
		})
	})

	Context("specify invalid range (end)", func() {
		It("should error", func() {
			c := context.WithValue(ctx, framework.CtxKey(idrange.VMIDRangeKey), "10-b")
			start, end, err := idrange.FindVMIDRange(c)
			Expect(start).To(Equal(0))
			Expect(end).To(Equal(0))
			Expect(err.Error()).To(ContainSubstring("invalid range is specified"))
		})
	})

	Context("specify valid range", func() {
		It("should not error", func() {
			c := context.WithValue(ctx, framework.CtxKey(idrange.VMIDRangeKey), "10-20")
			start, end, err := idrange.FindVMIDRange(c)
			Expect(start).To(Equal(10))
			Expect(end).To(Equal(20))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
