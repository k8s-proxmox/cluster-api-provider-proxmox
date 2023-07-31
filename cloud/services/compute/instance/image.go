package instance

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	rawImageDirPath = etcCAPPX + "/images"
)

// reconcileBootDevice
func (s *Service) reconcileBootDevice(ctx context.Context, vm *proxmox.VirtualMachine) error {
	log := log.FromContext(ctx)
	log.Info("reconcile boot device")

	// os image
	if err := s.setCloudImage(ctx); err != nil {
		return err
	}

	// volume
	if err := vm.ResizeVolume(ctx, bootDvice, s.scope.GetHardware().Disk); err != nil {
		return err
	}

	return nil
}

// setCloudImage downloads OS image into Proxmox node
// and then sets it to specified storage
func (s *Service) setCloudImage(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("setting cloud image")

	image := s.scope.GetImage()
	url := image.URL
	fileName := path.Base(url)
	rawImageFilePath := fmt.Sprintf("%s/%s", rawImageDirPath, fileName)

	// workaround
	// API does not support something equivalent of "qm importdisk"
	vnc, err := s.vncClient(s.scope.NodeName())
	defer vnc.Close()
	out, _, err := vnc.Exec(ctx, fmt.Sprintf("wget %s --directory-prefix %s -nc", url, rawImageDirPath))
	if err != nil {
		return errors.Errorf("failed to download image: %s : %v", out, err)
	}

	// checksum
	if image.Checksum != "" {
		cscmd, err := findValidChecksumCommand(*image.ChecksumType)
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf("echo -n '%s %s' | %s --check -", image.Checksum, rawImageFilePath, cscmd)
		out, _, err = vnc.Exec(context.TODO(), cmd)
		if err != nil {
			return errors.Errorf("failed to confirm checksum: %s : %v", out, err)
		}
	}

	vmid := s.scope.GetVMID()
	destPath := fmt.Sprintf("%s/images/%d/vm-%d-disk-0.raw", s.scope.GetStorage().Path, *vmid, *vmid)
	out, _, err = vnc.Exec(context.TODO(), fmt.Sprintf("/usr/bin/qemu-img convert -O raw %s %s", rawImageFilePath, destPath))
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
