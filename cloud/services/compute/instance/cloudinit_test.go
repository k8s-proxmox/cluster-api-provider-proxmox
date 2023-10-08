package instance_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/instance"
)

func TestCloudInit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CloudInit Suite")
}

var _ = Describe("mergeUserDatas", Label("unit", "cloudinit"), func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	var _ = Context("merge 3 user datas", func() {
		var _ = It("should no error", func() {
			a := infrav1.UserData{
				User:   "test-user",
				RunCmd: []string{"command C"},
			}
			b := infrav1.UserData{
				User:   "override-user",
				RunCmd: []string{"command A", "command B"},
			}
			c := infrav1.UserData{
				User:   "additional-user",
				RunCmd: []string{"command D"},
			}
			expected := infrav1.UserData{
				User:   "additional-user",
				RunCmd: []string{"command D", "command A", "command B", "command C"},
			}
			d, err := instance.MergeUserDatas(&a, &b, &c)
			Expect(err).NotTo(HaveOccurred())
			Expect(*d).To(Equal(expected))
		})
	})

	var _ = Context("merge 2 user datas and nil", func() {
		var _ = It("should no error", func() {
			a := infrav1.UserData{
				User:   "test-user",
				RunCmd: []string{"command C"},
			}
			b := infrav1.UserData{
				User:   "override-user",
				RunCmd: []string{"command A", "command B"},
			}
			expected := infrav1.UserData{
				User:   "override-user",
				RunCmd: []string{"command A", "command B", "command C"},
			}
			c, err := instance.MergeUserDatas(&a, &b, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(*c).To(Equal(expected))
		})
	})
})
