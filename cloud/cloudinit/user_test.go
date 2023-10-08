package cloudinit_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
)

func TestCloudInit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CloudInit Suite")
}

var _ = Describe("ParseUserDatas", Label("unit", "cloudinit"), func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("correct format", func() {
		It("should no error", func() {
			testYaml := `
write_files:
  - path: /run/kubeadm/kubeadm.yaml
    owner: root:root
    permissions: '0640'
    content: |
      asdfasdfasdf
runcmd:
  - 'kubeadm init --config /run/kubeadm/kubeadm.yaml  && echo success > /run/cluster-api/bootstrap-success.complete'
  - "curl -L https://dl.k8s.io/release/v1.27.3/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl"
  - "chmod +x /usr/local/bin/kubectl"
  - "reboot now"
  `
			userData, err := cloudinit.ParseUserData(testYaml)
			Expect(err).NotTo(HaveOccurred())
			Expect(userData).NotTo(BeNil())
		})
	})

	Context("incorrect format", func() {
		It("should error", func() {
			testYaml := `
write_files:
  - path: /run/kubeadm/kubeadm.yaml
owner: root:root
    permissions: '0640'
    content: |
      asdfasdfasdf
  `
			userData, err := cloudinit.ParseUserData(testYaml)
			Expect(err).To(HaveOccurred())
			Expect(userData).To(BeNil())
		})
	})
})

var _ = Describe("GenerateUserDataYaml", Label("unit", "cloudinit"), func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("generate user-data yaml string", func() {
		It("should no error", func() {
			userData := infrav1.UserData{
				RunCmd: []string{"echo", "pwd"},
			}
			yaml, err := cloudinit.GenerateUserDataYaml(userData)
			Expect(err).NotTo(HaveOccurred())
			Expect(yaml).NotTo(BeNil())
		})
	})
})

var _ = Describe("MergeUserDatas", Label("unit", "cloudinit"), func() {
	var _ = Context("merge 2 user datas", func() {
		var _ = It("should no error", func() {
			a := infrav1.UserData{
				User:   "override-user",
				RunCmd: []string{"command A", "command B"},
			}
			b := infrav1.UserData{
				User:   "test-user",
				RunCmd: []string{"command C"},
			}
			expected := infrav1.UserData{
				User:   "override-user",
				RunCmd: []string{"command A", "command B", "command C"},
			}

			c, err := cloudinit.MergeUserDatas(&a, &b)
			Expect(err).NotTo(HaveOccurred())
			Expect(*c).To(Equal(expected))
		})
	})
})
