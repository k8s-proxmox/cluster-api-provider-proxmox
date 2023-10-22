package storage_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/storage"
	"github.com/sp-yduck/cluster-api-provider-proxmox/internal/fake"
)

var _ = Describe("Delete", Label("integration", "storage"), func() {
	var service *storage.Service

	BeforeEach(func() {
		scope := fake.NewClusterScope(proxmoxSvc)
		service = storage.NewService(scope)
	})
	AfterEach(func() {})

	Context("Reconcile on Delete Storage", func() {
		It("should not error", func() {
			err := service.Delete(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Reconcile", Label("integration", "storage"), func() {
	var service *storage.Service

	BeforeEach(func() {
		scope := fake.NewClusterScope(proxmoxSvc)
		name := "cappx-integration-test"
		scope.SetStorage(infrav1.Storage{
			Name: name,
			Path: fmt.Sprintf("/tmp/cappx-test/%s", name),
		})
		service = storage.NewService(scope)
	})
	AfterEach(func() {
		err := service.Delete(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Reconcile Storage", func() {
		It("shold not error", func() {
			// there should be no storage yet
			err := service.Reconcile(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// there should be storage already
			err = service.Reconcile(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("generateVMStorageOptions", Label("unit", "storage"), func() {
	var scope *fake.FakeClusterScope

	BeforeEach(func() {
		scope = fake.NewClusterScope(nil)
	})
	AfterEach(func() {})

	Context("both name and path are specified", func() {
		It("option should inherit specified name and path", func() {
			testStorage := infrav1.Storage{
				Name: "foo",
				Path: "/bar/buz",
			}
			scope.SetStorage(testStorage)
			option := storage.GenerateVMStorageOptions(scope)
			Expect(option.Storage).To(Equal("foo"))
			Expect(option.Path).To(Equal("/bar/buz"))
		})
	})
})
