package instance

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

// reconcileBootDevice
func (s *Service) reconcileBootDevice(ctx context.Context, vm *vm.VirtualMachine) error {
	vmid := s.scope.GetVMID()
	storage := s.scope.GetStorage()
	image := s.scope.GetImage()
	hardware := s.scope.GetHardware()
	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("%v", hardware))

	// os image
	if err := SetCloudImage(ctx, *vmid, storage, image, s.remote); err != nil {
		return err
	}

	// volume
	if err := vm.ResizeVolume("scsi0", hardware.Disk); err != nil {
		return err
	}

	return nil
}

// setCloudImage downloads OS image into Proxmox node
// and then sets it to specified storage
func SetCloudImage(ctx context.Context, vmid int, storage infrav1.Storage, image infrav1.Image, ssh scope.SSHClient) error {
	log := log.FromContext(ctx)
	log.Info("setting cloud image")

	url := image.URL
	fileName := path.Base(url)
	rawImageDirPath := fmt.Sprintf("%s/images", etcCAPPX)
	rawImageFilePath := fmt.Sprintf("%s/%s", rawImageDirPath, fileName)

	// workaround
	// API does not support something equivalent of "qm importdisk"
	out, err := ssh.RunCommand(fmt.Sprintf("wget %s --directory-prefix %s -nc", url, rawImageDirPath))
	if err != nil {
		return errors.Errorf("failed to download image: %s : %v", out, err)
	}

	// checksum
	if image.Checksum != "" {
		cscmd, err := findValidChecksumCommand(*image.ChecksumType)
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf("%s --check -", cscmd)
		stdin := fmt.Sprintf("%s %s", image.Checksum, rawImageFilePath)
		out, err = ssh.RunWithStdin(cmd, stdin)
		if err != nil {
			return errors.Errorf("failed to confirm checksum: %s : %v", out, err)
		}
	}

	destPath := fmt.Sprintf("%s/images/%d/vm-%d-disk-0.raw", storage.Path, vmid, vmid)
	out, err = ssh.RunCommand(fmt.Sprintf("/usr/bin/qemu-img convert -O raw %s %s", rawImageFilePath, destPath))
	if err != nil {
		return errors.Errorf("failed to convert iamge : %s : %v", out, err)
	}
	return nil
}

func findValidChecksumCommand(csType string) (string, error) {
	csType = strings.ToLower(csType)
	switch csType {
	case "sha256", "sha256sum":
		return "sha256sum", nil
	case "md5", "md5sum":
		return "md5sum", nil
	default:
		return "", errors.Errorf("checksum type %s is not supported", csType)
	}
}
