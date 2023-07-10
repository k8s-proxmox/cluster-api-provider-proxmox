package instance

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
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

func baseUserData(vmName string) *infrav1.User {
	return &infrav1.User{
		HostName: vmName,
		Packages: []string{"qemu-guest-agent"},
		RunCmd:   []string{"systemctl start qemu-guest-agent"},
	}
}
