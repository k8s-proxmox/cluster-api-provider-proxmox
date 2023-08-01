package instance

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
)

const (
	userSnippetPathFormat = "snippets/%s-user.yml"
)

// reconcileCloudInit
func (s *Service) reconcileCloudInit(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling cloud init")

	// user
	if err := s.reconcileCloudInitUser(ctx); err != nil {
		return err
	}

	return nil
}

// delete CloudConfig
func (s *Service) deleteCloudConfig(ctx context.Context) error {
	storageName := s.scope.GetStorage().Name
	path := userSnippetPath(s.scope.Name())
	volumeID := fmt.Sprintf("%s:%s", storageName, path)

	node, err := s.client.Node(ctx, s.scope.NodeName())
	if err != nil {
		return err
	}
	storage, err := s.client.Storage(ctx, storageName)
	if err != nil {
		return err
	}
	storage.Node = node.Node
	return storage.DeleteVolume(ctx, volumeID)
}

func (s *Service) reconcileCloudInitUser(ctx context.Context) error {
	log := log.FromContext(ctx)

	// cloud init from bootstrap provider
	bootstrap, err := s.scope.GetBootstrapData()
	if err != nil {
		log.Error(err, "Error getting bootstrap data for machine")
		return errors.Wrap(err, "failed to retrieve bootstrap data")
	}
	bootstrapConfig, err := cloudinit.ParseUser(bootstrap)
	if err != nil {
		return err
	}

	vmName := s.scope.Name()
	cloudConfig, err := mergeUserDatas(bootstrapConfig, baseUserData(vmName), s.scope.GetCloudInit().User)
	if err != nil {
		return err
	}

	configYaml, err := cloudinit.GenerateUserYaml(*cloudConfig)
	if err != nil {
		return err
	}

	// to do: should be set via API
	vnc, err := s.vncClient(s.scope.NodeName())
	if err != nil {
		return err
	}
	defer vnc.Close()
	filePath := fmt.Sprintf("%s/%s", s.scope.GetStorage().Path, userSnippetPath(vmName))
	if err := vnc.WriteFile(context.TODO(), configYaml, filePath); err != nil {
		return errors.Errorf("failed to write file error : %v", err)
	}

	return nil
}

// a and b must not be nil
// only c can be nil
func mergeUserDatas(a, b, c *infrav1.User) (*infrav1.User, error) {
	var err error
	if c != nil {
		b, err = cloudinit.MergeUsers(b, c)
		if err != nil {
			return nil, err
		}
	}
	a, err = cloudinit.MergeUsers(a, b)
	if err != nil {
		return nil, err
	}
	return a, err
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
