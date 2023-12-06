package plugins_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/plugins"
)

func TestPlugins(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugins Suite")
}

var _ = Describe("GetPluginConfigFromFile", Label("unit", "scheduler"), func() {
	path := "./test-plugin-config.yaml"
	BeforeEach(func() {
		content := `scores:
  Random:
    enable: false`
		err := stringToFile(content, path)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := rm(path)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with empty file path", func() {
		path := ""
		It("should not error", func() {
			config, err := plugins.GetPluginConfigFromFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(plugins.PluginConfigs{}))
		})
	})

	Context("with non-empty file path", func() {
		It("should not error", func() {
			config, err := plugins.GetPluginConfigFromFile(path)
			Expect(err).NotTo(HaveOccurred())
			scores := map[string]plugins.PluginConfig{}
			scores["Random"] = plugins.PluginConfig{Enable: false}
			Expect(config).To(Equal(plugins.PluginConfigs{ScorePlugins: scores}))
		})
	})

	Context("with wrong file path", func() {
		It("shold error", func() {
			path := "./wrong-plugin-config.yaml"
			config, err := plugins.GetPluginConfigFromFile(path)
			Expect(err).To(HaveOccurred())
			Expect(config).To(Equal(plugins.PluginConfigs{}))
		})
	})
})

func stringToFile(str string, path string) error {
	b := []byte(str)
	return os.WriteFile(path, b, 0666)
}

func rm(path string) error {
	return os.Remove(path)
}
