package framework_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sp-yduck/proxmox-go/proxmox"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	proxmoxSvc *proxmox.Service
)

func TestFrameworks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scheduler Framework Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	if GinkgoLabelFilter() != "unit" {
		By("setup proxmox client to do integration test")
		url := os.Getenv("PROXMOX_URL")
		user := os.Getenv("PROXMOX_USER")
		password := os.Getenv("PROXMOX_PASSWORD")
		tokenid := os.Getenv("PROXMOX_TOKENID")
		secret := os.Getenv("PROXMOX_SECRET")

		authConfig := proxmox.AuthConfig{
			Username: user,
			Password: password,
			TokenID:  tokenid,
			Secret:   secret,
		}
		param := proxmox.NewParams(url, authConfig, proxmox.ClientConfig{InsecureSkipVerify: true})
		var err error
		proxmoxSvc, err = proxmox.GetOrCreateService(param)
		Expect(err).NotTo(HaveOccurred())
	}
})
