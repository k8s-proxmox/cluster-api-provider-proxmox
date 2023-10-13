package instance

import (
	"context"
	"fmt"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud"
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

	// volume
	if err := vm.ResizeVolume(ctx, bootDevice, s.scope.GetHardware().Disk); err != nil {
		return err
	}

	return nil
}

// importCloudImage downloads OS image into Proxmox node, converts it to qcow2 format and returns proxmox "import-from"
// compatible string
func (s *Service) importCloudImage(ctx context.Context) (string, error) {
	log := log.FromContext(ctx)
	log.Info("importing cloud image")

	image := s.scope.GetImage()
	clusterStorageImagesBasePath := clusterStorageImagesBasePath(s.scope)
	sourceImageFilePath := sourceImagePath(s.scope)
	gcow2ImageFilePath := qcow2ImagePath(sourceImageFilePath)

	// workaround
	// API does not support something equivalent of "qm importdisk"
	vnc, err := s.vncClient(*s.scope.NodeName())
	if err != nil {
		return "", errors.Errorf("failed to create vnc client: %v", err)
	}
	defer vnc.Close()

	gcow2Present, _ := isFilePresent(vnc, gcow2ImageFilePath)
	if !gcow2Present {

		sourceImagePresent, _ := isFilePresent(vnc, sourceImageFilePath)
		if !sourceImagePresent {
			out, _, err := vnc.Exec(ctx, fmt.Sprintf("mkdir -p %s", clusterStorageImagesBasePath))
			if err != nil {
				return "", errors.Errorf("failed to create dir %s: %s : %v", clusterStorageImagesBasePath, out, err)
			}
			log.Info("downloading node image. this will take few mins.")
			out, _, err = vnc.Exec(ctx, fmt.Sprintf("wget '%s' -O '%s'", image.URL, sourceImageFilePath))
			if err != nil {
				return "", errors.Errorf("failed to download image: %s : %v", out, err)
			}
			log.Info("node image downloaded")
		}

		if _, err = isChecksumOK(vnc, image, sourceImageFilePath); err != nil {
			if _, err = deleteFile(vnc, sourceImageFilePath); err != nil {
				return "", errors.Errorf("failed to delete source image after checksum failed: %v", err)
			} else {
				return "", errors.Errorf("failed to confirm checksum: %v", err)
			}
		}
		log.Info("node image downloaded")

		out, _, err := vnc.Exec(context.TODO(), fmt.Sprintf("/usr/bin/qemu-img convert -O qcow2 '%s' '%s'", sourceImageFilePath, gcow2ImageFilePath))
		if err != nil {
			return "", errors.Errorf("failed to convert image : %s : %v", out, err)
		}
		log.Info("converted node image now available")

		if _, err = deleteFile(vnc, sourceImageFilePath); err != nil {
			log.Info("failed to delete source image after conversion, ignoring error")
		}
	}

	// convert absolute path <cluster storage path>/images/0/<filename> to <storage name>:0/<filename>
	return strings.Replace(gcow2ImageFilePath, clusterStorageImagesBasePath, s.scope.GetClusterStorage().Name+":0", 1), nil
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

func isFilePresent(client *proxmox.VNCWebSocketClient, path string) (bool, error) {

	cmd := fmt.Sprintf("test -f '%s'", path)
	out, _, err := client.Exec(context.TODO(), cmd)
	if err != nil {
		return false, errors.Errorf("failed to find file: %s : %v", out, err)
	}

	return true, nil
}

func deleteFile(client *proxmox.VNCWebSocketClient, filePath string) (bool, error) {
	cmd := fmt.Sprintf("rm '%s'", filePath)
	out, _, err := client.Exec(context.TODO(), cmd)
	if err != nil {
		return false, errors.Errorf("failed to delete file: %s : %v", out, err)
	}

	return true, nil
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

func clusterStorageImagesBasePath(scope cloud.MachineGetter) string {
	// we are using 0 as vm id to workaround limitation of import-from that expects vm id disks
	return fmt.Sprintf("%s/images/0", scope.GetClusterStorage().Path)
}

func sourceImagePath(scope cloud.MachineGetter) string {
	image := scope.GetImage()

	fileName := path.Base(image.URL)
	if image.Checksum != "" {
		fileName = image.Checksum + "." + fileName
	}

	return fmt.Sprintf("%s/%s", clusterStorageImagesBasePath(scope), fileName)
}

func qcow2ImagePath(sourceImageFilePath string) string {
	ext := path.Base(path.Ext(sourceImageFilePath))
	return strings.Replace(sourceImageFilePath, ext, ".qcow2", 1)
}
