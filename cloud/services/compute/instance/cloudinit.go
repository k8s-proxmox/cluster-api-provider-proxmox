package instance

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

const (
	userSnippetPathFormat = "snippets/%s-user.yml"
)

// reconcileCloudInit
func (s *Service) reconcileCloudInit(bootstrap string) error {
	// user
	if err := s.reconcileCloudInitUser(bootstrap); err != nil {
		return err
	}
	return nil
}

// delete CloudConfig
func (s *Service) deleteCloudConfig() error {
	storageName := s.scope.GetStorage().Name
	path := userSnippetPath(s.scope.Name())
	volumeID := fmt.Sprintf("%s:%s", storageName, path)

	node, err := s.client.Node(s.scope.NodeName())
	if err != nil {
		return err
	}
	storage, err := node.Storage(storageName)
	if err != nil {
		return err
	}
	content, err := storage.GetContent(volumeID)
	if IsNotFound(err) { // return nil if it's already deleted
		return nil
	}
	if err != nil {
		return err
	}

	return content.DeleteVolume()
}

func (s *Service) reconcileCloudInitUser(bootstrap string) error {
	vmName := s.scope.Name()
	storagePath := s.scope.GetStorage().Path
	config := s.scope.GetCloudInit().User

	bootstrapConfig, err := cloudinit.ParseUser(bootstrap)
	if err != nil {
		return err
	}
	base := baseUserData(vmName)
	if config != nil {
		base, err = cloudinit.MergeUsers(*config, *base)
		if err != nil {
			return err
		}
	}
	cloudConfig, err := cloudinit.MergeUsers(*base, *bootstrapConfig)
	if err != nil {
		return err
	}
	configYaml, err := cloudinit.GenerateUserYaml(*cloudConfig)
	if err != nil {
		return err
	}

	klog.Info(configYaml)

	// to do: should be set via API
	out, err := s.remote.RunWithStdin(fmt.Sprintf("tee %s/%s", storagePath, userSnippetPath(vmName)), configYaml)
	if err != nil {
		return errors.Errorf("ssh command error : %s : %v", out, err)
	}

	return nil
}

func userSnippetPath(vmName string) string {
	return fmt.Sprintf(userSnippetPathFormat, vmName)
}

// DEPRECATED : cicustom should be set via API
func ApplyCICustom(vmid int, vmName, storageName, ciType string, ssh scope.SSHClient) error {
	if !cloudinit.IsValidType(ciType) {
		return errors.Errorf("invalid cloud init type: %s", ciType)
	}
	cicustom := fmt.Sprintf("%s=%s:snippets/%s-%s.yml", ciType, storageName, vmName, ciType)
	out, err := ssh.RunCommand(fmt.Sprintf("qm set %d --cicustom '%s'", vmid, cicustom))
	if err != nil {
		return errors.Errorf("ssh command error : %s : %v", out, err)
	}
	return nil
}

// to do : remove these cloud-config
func baseUserData(vmName string) *infrav1.User {
	return &infrav1.User{
		GrowPart:       infrav1.GrowPart{Mode: "auto", Devices: []string{"/"}, IgnoreGrowrootDisabled: false},
		HostName:       vmName,
		ManageEtcHosts: true,
		ChPasswd:       infrav1.ChPasswd{Expire: "False"},
		Users:          []string{"default"},
		Packages:       []string{"socat", "conntrack"},
		PackageUpgrade: true,
		WriteFiles: []infrav1.WriteFiles{
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
		},
		RunCmd: []string{
			"modprobe overlay",
			"modprobe br_netfilter",
			"sysctl --system",
			`mkdir -p /usr/local/bin`,
			`curl -L "https://github.com/containerd/containerd/releases/download/v1.7.2/containerd-1.7.2-linux-amd64.tar.gz" | tar Cxvz "/usr/local"`,
			`curl -L "https://raw.githubusercontent.com/containerd/containerd/main/containerd.service" -o /etc/systemd/system/containerd.service`,
			"mkdir -p /etc/containerd",
			"containerd config default > /etc/containerd/config.toml",
			"sed 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml -i",
			"systemctl daemon-reload",
			"systemctl enable --now containerd",
			"mkdir -p /usr/local/sbin",
			`curl -L "https://github.com/opencontainers/runc/releases/download/v1.1.7/runc.amd64" -o /usr/local/sbin/runc`,
			"chmod 755 /usr/local/sbin/runc",
			"mkdir -p /opt/cni/bin",
			`curl -L "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz" | tar -C "/opt/cni/bin" -xz`,
			`curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz" | tar -C "/usr/local/bin" -xz`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.26.5/bin/linux/amd64/kubeadm -o /usr/local/bin/kubeadm`,
			`chmod +x /usr/local/bin/kubeadm`,
			`curl -L --remote-name-all https://dl.k8s.io/release/v1.26.5/bin/linux/amd64/kubelet -o /usr/local/bin/kubelet`,
			`chmod +x /usr/local/bin/kubelet`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service`,
			`mkdir -p /etc/systemd/system/kubelet.service.d`,
			`curl -sSL "https://raw.githubusercontent.com/kubernetes/release/v0.15.1/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf" | sed "s:/usr/bin:/usr/local/bin:g" | tee /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`,
			"systemctl enable kubelet.service",
		},
	}
}
