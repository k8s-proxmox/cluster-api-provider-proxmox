package framework_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
)

var _ = Describe("ContextWithMap", Label("unit", "framework"), func() {
	c := context.Background()

	Context("input empty map", func() {
		It("should not panic", func() {
			m := map[string]string{}
			ctx := framework.ContextWithMap(c, m)
			Expect(ctx).To(Equal(c))
		})
	})

	Context("input nil", func() {
		It("should not panic", func() {
			ctx := framework.ContextWithMap(c, nil)
			Expect(ctx).To(Equal(c))
		})
	})

	Context("input random map", func() {
		It("should get map's key-value", func() {
			m := map[string]string{}
			m["abc"] = "ABC"
			ctx := framework.ContextWithMap(c, m)
			Expect(ctx.Value("abc")).ToNot(Equal("ABC"))
			Expect(ctx.Value(framework.CtxKey("abc"))).To(Equal("ABC"))
		})
	})
})
