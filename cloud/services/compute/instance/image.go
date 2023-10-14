package instance

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
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
func (s *Service) setCloudImage(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("setting cloud image")

	image := s.scope.GetImage()
	rawImageFilePath := rawImageFilePath(image)

	vnc, err := s.vncClient(s.scope.NodeName())
	if err != nil {
		return errors.Errorf("failed to create vnc client: %v", err)
	}
	defer vnc.Close()

	// download image
	ok, _ := isChecksumOK(vnc, image, rawImageFilePath)
	if !ok { // if checksum is ok, it means the image is already there. skip installing
		out, _, err := vnc.Exec(ctx, fmt.Sprintf("mkdir -p %s && mkdir -p %s", etcCAPPX, rawImageDirPath))
		if err != nil {
			return errors.Errorf("failed to create dir %s: %s : %v", rawImageDirPath, out, err)
		}
		log.Info("downloading node image. this will take few mins.")
		out, _, err = vnc.Exec(ctx, fmt.Sprintf("wget %s -O %s", image.URL, rawImageFilePath))
		if err != nil {
			return errors.Errorf("failed to download image: %s : %v", out, err)
		}
		if _, err = isChecksumOK(vnc, image, rawImageFilePath); err != nil {
			return errors.Errorf("failed to confirm checksum: %v", err)
		}
	}

	// // convert downloaded image to raw format and set it to storage
	// vmid := s.scope.GetVMID()
	// destPath := fmt.Sprintf("%s/images/%d/vm-%d-disk-0.raw", s.scope.GetStorage().Path, *vmid, *vmid)
	// out, _, err := vnc.Exec(context.TODO(), fmt.Sprintf("/usr/bin/qemu-img convert -O raw %s %s", rawImageFilePath, destPath))
	// if err != nil {
	// 	return errors.Errorf("failed to convert iamge : %s : %v", out, err)
	// }

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

func isChecksumOK(client *proxmox.VNCWebSocketClient, image infrav1.Image, path string) (bool, error) {
	if image.Checksum != "" {
		cscmd, err := findValidChecksumCommand(*image.ChecksumType)
		if err != nil {
			return false, err
		}
		cmd := fmt.Sprintf("echo -n '%s %s' | %s --check -", image.Checksum, path, cscmd)
		out, _, err := client.Exec(context.TODO(), cmd)
		if err != nil {
			return false, errors.Errorf("failed to confirm checksum: %s : %v", out, err)
		}
		return true, nil
	}
	return false, nil
}

func rawImageFilePath(image infrav1.Image) string {
	fileName := path.Base(image.URL)
	if image.Checksum != "" {
		fileName = image.Checksum + "." + fileName
	}
	return fmt.Sprintf("%s/%s", rawImageDirPath, fileName)
}
