package resourcepool_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/resourcepool"
	"github.com/sp-yduck/cluster-api-provider-proxmox/internal/fake"
)

var _ = Describe("Delete", Label("integration", "resourcepool"), func() {
	var service *resourcepool.Service

	BeforeEach(func() {
		scope := fake.NewClusterScope(proxmoxSvc)
		service = resourcepool.NewService(scope)
		err := service.Reconcile(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {})

	Context("Reconcile on Delete resource pool", func() {
		It("should not error", func() {

			// there should be pool
			err := service.Delete(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// there should be no pool already
			err = service.Delete(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Reconcile", Label("integration", "resourcepool"), func() {
	var service *resourcepool.Service

	BeforeEach(func() {
		scope := fake.NewClusterScope(proxmoxSvc)
		service = resourcepool.NewService(scope)
	})
	AfterEach(func() {
		err := service.Delete(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Reconcile resourcepool", func() {
		It("shold not error", func() {
			// there should be no resourcepool yet
			err := service.Reconcile(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// there should be resourcepool already
			err = service.Reconcile(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
