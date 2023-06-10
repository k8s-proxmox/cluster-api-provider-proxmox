package instance

import (
	"fmt"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type UserConfig struct {
	HostName       string       `yaml:"hostname,omitempty"`
	ManageEtcHosts bool         `yaml:"manage_etc_hosts,omitempty"`
	User           string       `yaml:"user,omitempty"`
	ChPasswd       ChPasswd     `yaml:"chpasswd,omitempty"`
	Users          []string     `yaml:"users,omitempty"`
	Password       string       `yaml:"password,omitempty"`
	PackageUpgrade bool         `yaml:"package_upgrade,omitempty"`
	WriteFiles     []WriteFiles `yaml:"write_files,omitempty"`
	RunCmd         []string     `yaml:"runcmd,omitempty"`
}

type ChPasswd struct {
	Expire string `yaml:"expire,omitempty"`
}

type WriteFiles struct {
	Path        string `yaml:"path,omitempty"`
	Owner       string `yaml:"owner,omitempty"`
	Permissions string `yaml:"permissions,omitempty"`
	Content     string `yaml:"content,omitempty"`
}

func ParseUserConfig(content string) UserConfig {
	var config *UserConfig
	yaml.Unmarshal([]byte(content), &config)
	return *config
}

func GenerateUserConfigYaml(config UserConfig) (string, error) {
	b, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#cloud-config\n%s", string(b)), nil
}

func MergeUserConfigs(a, b UserConfig) (*UserConfig, error) {
	if err := mergo.Merge(&a, b); err != nil {
		return nil, err
	}
	return &a, nil
}

func baseUserConfig(vmName string) UserConfig {
	return UserConfig{
		HostName:       vmName,
		ManageEtcHosts: true,
		User:           "ubnt", // to do
		Password:       "ubnt", // to do
		ChPasswd:       ChPasswd{Expire: "False"},
		Users:          []string{"default"},
		PackageUpgrade: true,
		RunCmd: []string{
			"mkdir -p /opt/cni/bin",
			`curl -L "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz" | tar -C "/opt/cni/bin" -xz`,
			`mkdir -p /usr/local/bin`,
			`curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz" | tar -C "/usr/local/bin" -xz`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.22.2/bin/linux/amd64/kubeadm -o /usr/local/bin/kubeadm`,
			`chmod +x /usr/local/bin/kubeadm`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.22.2/bin/linux/amd64/kubelet -o /usr/local/bin/kubelet`,
			`chmod +x /usr/local/bin/kubelet`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service`,
			`mkdir -p /etc/systemd/system/kubelet.service.d`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`,
		},
	}
}

// func generateBaseCloudConfig(vmName string) string {
// 	// do not use tab
// 	return fmt.Sprintf(`#cloud-config
// hostname: %s
// manage_etc_hosts: true
// user: ubnt
// password: ubnt
// chpasswd:
//   expire: False
// users:
//   - default
// package_upgrade: true
// runcmd:
// - mkdir -p /opt/cni/bin
// - curl -L "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz" | tar -C "/opt/cni/bin" -xz
// - mkdir -p /usr/local/bin
// - curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz" | tar -C "/usr/local/bin" -xz
// - curl -L --remote-name-all https://dl.k8s.io/release/v1.27.2/bin/linux/amd64/kubeadm -o /usr/local/bin/kubeadm
// - chmod +x /usr/local/bin/kubeadm
// - curl -L --remote-name-all https://dl.k8s.io/release/v1.27.2/bin/linux/amd64/kubelet -o /usr/local/bin/kubelet
// - chmod +x /usr/local/bin/kubelet
// - curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service
// - mkdir -p /etc/systemd/system/kubelet.service.d
// - curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`, vmName)

// }

// func removeComments(content string) string {
// 	content = strings.Replace(content, "#cloud-config\n", "", 1)
// 	content = strings.Replace(content, "## template: jinja", "", 1)
// 	return strings.Replace(content, "\n\n", "\n", 1)
// }

// func mergeYamls(a, b string) (string, error) {

// 	var master map[string]interface{}
// 	bs := []byte(a)
// 	if err := yaml.Unmarshal(bs, &master); err != nil {
// 		return "", err
// 	}

// 	var override map[string]interface{}
// 	bs = []byte(b)
// 	if err := yaml.Unmarshal(bs, &override); err != nil {
// 		return "", err
// 	}

// 	for k, v := range override {
// 		master[k] = v
// 	}

// 	bs, err := yaml.Marshal(master)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(bs), nil
// }
