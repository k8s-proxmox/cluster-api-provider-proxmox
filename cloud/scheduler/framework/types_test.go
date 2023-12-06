package framework_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
)

var _ = Describe("GetNodeInfoList", Label("integration", "framework"), func() {
	ctx := context.Background()

	It("should not error", func() {
		nodes, err := framework.GetNodeInfoList(ctx, proxmoxSvc)
		Expect(err).To(BeNil())
		Expect(len(nodes)).ToNot(Equal(0))
	})

})
