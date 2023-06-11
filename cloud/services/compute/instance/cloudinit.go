package instance

import (
	"fmt"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type UserConfig struct {
	GrowPart       GrowPart     `yaml:"growpart,omitempty"`
	HostName       string       `yaml:"hostname,omitempty"`
	ManageEtcHosts bool         `yaml:"manage_etc_hosts,omitempty"`
	User           string       `yaml:"user,omitempty"`
	ChPasswd       ChPasswd     `yaml:"chpasswd,omitempty"`
	Users          []string     `yaml:"users,omitempty"`
	Password       string       `yaml:"password,omitempty"`
	Packages       []string     `yaml:"packages,omitempty"`
	PackageUpgrade bool         `yaml:"package_upgrade,omitempty"`
	WriteFiles     []WriteFiles `yaml:"write_files,omitempty"`
	RunCmd         []string     `yaml:"runcmd,omitempty"`
}

type GrowPart struct {
	Mode                   string   `yaml:"mode,omitempty"`
	Devices                []string `yaml:"devices,omitempty"`
	IgnoreGrowrootDisabled bool     `yaml:"ignore_growroot_disabled,omitempty"`
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
	if err := mergo.Merge(&a, b, mergo.WithAppendSlice); err != nil {
		return nil, err
	}
	return &a, nil
}

func baseUserConfig(vmName string) UserConfig {
	return UserConfig{
		GrowPart:       GrowPart{Mode: "auto", Devices: []string{"/"}, IgnoreGrowrootDisabled: false},
		HostName:       vmName,
		ManageEtcHosts: true,
		User:           "ubnt", // to do
		Password:       "ubnt", // to do
		ChPasswd:       ChPasswd{Expire: "False"},
		Users:          []string{"default"},
		Packages:       []string{"socat", "conntrack"},
		PackageUpgrade: true,
		WriteFiles: []WriteFiles{
			{
				Path:        "/etc/modules-load.d/k8s.conf",
				Owner:       "root:root",
				Permissions: "0640",
				Content:     "overlay\nbr_netfilter",
			},
			{
				Path:        "/etc/sysctl.d/k8s.conf",
				Owner:       "root:root",
				Permissions: "0640",
				Content: `net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1`,
			},
			{
				Path:        "/etc/containerd/config.toml",
				Owner:       "root:root",
				Permissions: "0640",
				Content: `[plugins."io.containerd.grpc.v1.cri"]
  sandbox_image = "registry.k8s.io/pause:3.2"
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
    SystemdCgroup = true`,
			},
		},
		RunCmd: []string{
			"modprobe overlay",
			"modprobe br_netfilter",
			"sysctl --system",
			`mkdir -p /usr/local/bin`,
			`curl -L "https://github.com/containerd/containerd/releases/download/v1.7.2/containerd-1.7.2-linux-amd64.tar.gz" | tar Cxvz "/usr/local"`,
			`curl -L "https://raw.githubusercontent.com/containerd/containerd/main/containerd.service" -o /etc/systemd/system/containerd.service`,
			"systemctl daemon-reload",
			"systemctl enable --now containerd",
			"mkdir -p /usr/local/sbin",
			`curl -L "https://github.com/opencontainers/runc/releases/download/v1.1.7/runc.amd64" -o /usr/local/sbin/runc`,
			"chmod 755 /usr/local/sbin/runc",
			"mkdir -p /opt/cni/bin",
			`curl -L "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz" | tar -C "/opt/cni/bin" -xz`,
			`curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz" | tar -C "/usr/local/bin" -xz`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.22.2/bin/linux/amd64/kubeadm -o /usr/local/bin/kubeadm`,
			`chmod +x /usr/local/bin/kubeadm`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.21.2/bin/linux/amd64/kubelet -o /usr/local/bin/kubelet`,
			`chmod +x /usr/local/bin/kubelet`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service`,
			`mkdir -p /etc/systemd/system/kubelet.service.d`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`,
			"systemctl enable kubelet.service",
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
