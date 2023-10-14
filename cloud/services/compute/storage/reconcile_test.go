package storage_test

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/storage"
	"github.com/sp-yduck/cluster-api-provider-proxmox/internal/fake"
)

var _ = Describe("Delete", Label("integration", "storage"), func() {
	var service *storage.Service

	BeforeEach(func() {
		scope := fake.NewMachineScope(proxmoxSvc)
		node, err := getRandomNode(scope)
		Expect(err).ToNot(HaveOccurred())
		scope.SetNodeName(node)
		name := "cappx-integration-test"
		scope.SetSnippetStorage(infrav1.SnippetStorage{
			Name: "dir-" + name,
			Path: fmt.Sprintf("/tmp/cappx-test/%s", name),
		})
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
		scope := fake.NewMachineScope(proxmoxSvc)
		node, err := getRandomNode(scope)
		Expect(err).ToNot(HaveOccurred())
		scope.SetNodeName(node)
		name := "cappx-integration-test"
		scope.SetSnippetStorage(infrav1.SnippetStorage{
			Name: "dir-" + name,
			Path: fmt.Sprintf("/tmp/cappx-test/%s", name),
		})
		scope.SetImageStorage(infrav1.ImageStorage{
			Name: os.Getenv("PROXMOX_IMAGE_STORAGE"),
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

var _ = Describe("generateSnippetStorageOptions", Label("unit", "storage"), func() {
	var scope *fake.FakeMachineScope

	BeforeEach(func() {
		scope = fake.NewMachineScope(nil)
	})
	AfterEach(func() {})

	Context("both name and path are specified", func() {
		It("option should inherit specified name and path", func() {
			testStorage := infrav1.SnippetStorage{
				Name: "foo",
				Path: "/bar/buz",
			}
			scope.SetSnippetStorage(testStorage)
			option := storage.GenerateSnippetStorageOptions(scope)
			Expect(option.Storage).To(Equal("foo"))
			Expect(option.Path).To(Equal("/bar/buz"))
		})
	})

	Context("both name and path are NOT specified", func() {
		It("option should have default name and path", func() {
			scope.SetName("foo-cluster")
			scope.SetSnippetStorage(infrav1.SnippetStorage{})
			option := storage.GenerateSnippetStorageOptions(scope)
			Expect(option.Storage).To(Equal("local-dir-foo-cluster"))
			Expect(option.Path).To(Equal("/var/lib/vz/local-dir-foo-cluster"))
		})
	})
})

func getRandomNode(scope storage.Scope) (string, error) {
	client := scope.CloudClient()
	nodes, err := client.Nodes(context.Background())
	if err != nil {
		return "", err
	}
	return nodes[0].Node, nil
}
