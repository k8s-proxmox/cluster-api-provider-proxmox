package providerid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/providerid"
)

func TestProviderID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProviderID Suite")
}

var _ = Describe("New", Label("unit", "providerid"), func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("empty uuid", func() {
		It("should error", func() {
			uuid := ""
			_, err := providerid.New(uuid)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("non empty uuid", func() {
		It("should not error", func() {
			uuid := "asdf"
			providerID, err := providerid.New(uuid)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID.UUID()).To(Equal("asdf"))
			Expect(providerID.String()).To(Equal("proxmox://asdf"))
		})
	})
})

var _ = Describe("UUID method", Label("unit", "providerid"), func() {
	pid, err := providerid.New("asdf")
	Expect(err).ToNot(HaveOccurred())

	BeforeEach(func() {})
	AfterEach(func() {})

	Context("uuid", func() {
		It("should not error", func() {
			uuid := pid.UUID()
			Expect(uuid).To(Equal("asdf"))
		})
	})
})

var _ = Describe("String method", Label("unit", "providerid"), func() {
	pid, err := providerid.New("asdf")
	Expect(err).ToNot(HaveOccurred())

	BeforeEach(func() {})
	AfterEach(func() {})

	Context("string", func() {
		It("should not error", func() {
			providerID := pid.String()
			Expect(providerID).To(Equal("proxmox://asdf"))
		})
	})
})
