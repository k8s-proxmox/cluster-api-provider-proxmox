package cloudinit_test

import (
	"testing"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
)

func TestParseUser(t *testing.T) {
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
	_, err := cloudinit.ParseUser(testYaml)
	if err != nil {
		t.Fatalf("failed to parse user: %v", err)
	}
}

func TestGenerateUserYaml(t *testing.T) {
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

	uc, err := cloudinit.ParseUser(testYaml)
	if err != nil {
		t.Fatalf("failed to parse user: %v", err)
	}

	_, err = cloudinit.GenerateUserYaml(*uc)
	if err != nil {
		t.Fatalf("generate : %v", err)
	}
}
