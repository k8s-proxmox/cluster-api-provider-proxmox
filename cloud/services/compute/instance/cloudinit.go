package instance

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

// reconcileCloudInit
func reconcileCloudInit(s *Service, vmid int, bootstrap string) error {
	vmName := s.scope.Name()
	storageName := s.scope.GetStorage().Name
	cloudInit := s.scope.GetCloudInit()

	klog.Info(cloudInit)

	// user
	if err := reconcileCloudInitUser(vmid, vmName, storageName, bootstrap, cloudInit.User, s.remote); err != nil {
		return err
	}

	// meta & network
	if err := reconcileCloudInitConfig(vmid, vmName, storageName, cloudInit, s.remote); err != nil {
		return err
	}

	return nil
}

func reconcileCloudInitUser(vmid int, vmName, storageName, bootstrap string, config infrav1.User, ssh scope.SSHClient) error {
	base := baseUserData(vmName)

	bootstrapConfig, err := cloudinit.ParseUser(bootstrap)
	if err != nil {
		return err
	}
	additional, err := cloudinit.MergeUsers(config, base)
	if err != nil {
		return err
	}
	cloudConfig, err := cloudinit.MergeUsers(*additional, *bootstrapConfig)
	if err != nil {
		return err
	}
	configYaml, err := cloudinit.GenerateUserYaml(*cloudConfig)
	if err != nil {
		return err
	}

	// to do: should be set via API
	out, err := ssh.RunWithStdin(fmt.Sprintf("tee /var/lib/vz/%s/snippets/%s-user.yml", storageName, vmName), configYaml)
	if err != nil {
		return errors.Errorf("ssh command error : %s : %v", out, err)
	}

	if err := ApplyCICustom(vmid, vmName, storageName, "user", ssh); err != nil {
		return err
	}
	return nil
}

func reconcileCloudInitConfig(vmid int, vmName, storageName string, cloudInit infrav1.CloudInit, ssh scope.SSHClient) error {
	klog.Info(cloudInit.Network)

	networkYaml, err := cloudinit.GenerateNetworkYaml(cloudInit.Network)
	if err != nil {
		return err
	}
	out, err := ssh.RunWithStdin(fmt.Sprintf("tee /var/lib/vz/%s/snippets/%s-network.yml", storageName, vmName), networkYaml)
	if err != nil {
		return errors.Errorf("ssh command error : %s : %v", out, err)
	}
	if err := ApplyCICustom(vmid, vmName, storageName, "network", ssh); err != nil {
		return err
	}

	// if meta != nil {
	// 	metaYaml, err := cloudinit.GenerateMetaYaml(*meta)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	out, err := ssh.RunWithStdin(fmt.Sprintf("tee /var/lib/vz/%s/snippets/%s-meta.yml", storageName, vmName), metaYaml)
	// 	if err != nil {
	// 		return errors.Errorf("ssh command error : %s : %v", out, err)
	// 	}
	// 	if err := ApplyCICustom(vmid, vmName, storageName, "meta", ssh); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

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

func baseUserData(vmName string) infrav1.User {
	return infrav1.User{
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
