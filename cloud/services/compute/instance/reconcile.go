package instance

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling instance resources")
	instance, err := s.createOrGetInstance(ctx)
	if err != nil {
		log.Error(err, "failed to create/get instance")
		return err
	}
	log.Info(fmt.Sprintf("instance : %v", instance))

	s.scope.SetProviderID(instance)
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))
	return nil
}

func (s *Service) createOrGetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	log.Info("Getting bootstrap data for machine")
	bootstrapData, err := s.scope.GetBootstrapData()
	if err != nil {
		log.Error(err, "Error getting bootstrap data for machine")
		return nil, errors.Wrap(err, "failed to retrieve bootstrap data")
	}

	if s.scope.GetInstanceID() == nil {
		log.Info("ProxmoxMachine doesn't have instanceID. instance will be created")
		return s.CreateInstance(ctx, bootstrapData)
	}
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if IsNotFound(err) {
			log.Info("instance wasn't found. new instance will be created")
			return s.CreateInstance(ctx, bootstrapData)
		}
		log.Error(err, "failed to get instance")
		return nil, err
	}
	return instance, nil
}

func (s *Service) GetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)
	instanceID := s.scope.GetInstanceID()
	vm, err := s.getInstanceFromInstanceID(*instanceID)
	if err != nil {
		if api.IsNotFound(err) {
			log.Info("instance wasn't found")
			return nil, api.ErrNotFound
		}
		log.Error(err, "failed to get instance from instance ID")
		return nil, err
	}
	return vm, nil
}

func (s *Service) getInstanceFromInstanceID(instanceID string) (*vm.VirtualMachine, error) {
	vmid, err := strconv.Atoi(instanceID)
	if err != nil {
		return nil, err
	}
	nodes, err := s.client.Nodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("proxmox nodes not found")
	}
	for _, node := range nodes {
		vm, err := node.VirtualMachine(vmid)
		if err != nil {
			continue
		}
		return vm, nil
	}
	return nil, api.ErrNotFound
}

func (s *Service) CreateInstance(ctx context.Context, bootstrap string) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// temp solution
	node, err := s.GetRandomNode()
	if err != nil {
		log.Error(err, "failed to get random node")
		return nil, err
	}

	vmid, err := s.GetNextID()
	if err != nil {
		log.Error(err, "failed to get available vmid")
		return nil, err
	}

	// create vm
	vmoption := generateVMOptions(s.scope.Name(), s.scope.GetStorage().Name, s.scope.GetNetwork())
	vm, err := node.CreateVirtualMachine(vmid, vmoption)
	if err != nil {
		log.Error(err, "failed to create virtual machine")
		return nil, err
	}

	// cloud init
	if err := reconcileCloudInit(s, vmid, bootstrap); err != nil {
		return nil, err
	}

	// os image
	if err := SetCloudImage(ctx, vmid, s.scope.GetStorage().Name, s.remote); err != nil {
		return nil, err
	}

	// volume
	if err := vm.ResizeVolume("scsi0", "+30G"); err != nil {
		return nil, err
	}

	// vm status
	if err := EnsureRunning(*vm); err != nil {
		return nil, err
	}
	return vm, nil
}

func IsNotFound(err error) bool {
	return api.IsNotFound(err)
}

func (s *Service) GetNextID() (int, error) {
	return s.client.NextID()
}

func (s *Service) GetNodes() ([]*node.Node, error) {
	return s.client.Nodes()
}

// GetRandomNode returns a node chosen randomly
func (s *Service) GetRandomNode() (*node.Node, error) {
	nodes, err := s.GetNodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) <= 0 {
		return nil, errors.Errorf("no nodes found")
	}
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	return nodes[r.Intn(len(nodes))], nil
}

// wip
func (s *Service) Delete(ctx context.Context) error {
	instance, err := s.GetInstance(ctx)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		return nil
	}
	// to do : stop instance before deletion
	return instance.Delete()
}

func SetCloudImage(ctx context.Context, vmid int, storageName string, ssh scope.SSHClient) error {
	log := log.FromContext(ctx)
	log.Info("setting cloud image")

	// workaround
	// API does not support something equivalent of "qm importdisk"
	out, err := ssh.RunCommand(fmt.Sprintf("wget https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64-disk-kvm.img -O /etc/capi-proxmox/jammy-server-cloudimg-amd64-disk-kvm.img -nc"))
	// if err != nil {
	// 	return nil, errors.Errorf("failed to download image")
	// }

	destPath := fmt.Sprintf("/var/lib/vz/%s/images/%d/vm-%d-disk-0.raw", storageName, vmid, vmid)
	out, err = ssh.RunCommand(fmt.Sprintf("/usr/bin/qemu-img convert -O raw /root/jammy-server-cloudimg-amd64-disk-kvm.img %s", destPath))
	if err != nil {
		return err
	}
	log.Info("imported cloud image")
	log.Info(out)
	return nil
}

func generateVMOptions(vmName, storageName string, network infrav1.Network) vm.VirtualMachineCreateOptions {

	vmoptions := vm.VirtualMachineCreateOptions{
		Agent:        "enabled=1",
		Cores:        2,
		Memory:       1024 * 4,
		Name:         vmName,
		NameServer:   network.NameServer,
		Boot:         "order=scsi0",
		Ide:          vm.Ide{Ide2: fmt.Sprintf("file=%s:cloudinit,media=cdrom", storageName)},
		CiCustom:     fmt.Sprintf("user=%s:snippets/%s-user.yml", storageName, vmName),
		IPConfig:     vm.IPConfig{IPConfig0: network.IPConfig.String()},
		OSType:       vm.L26,
		Net:          vm.Net{Net0: "model=virtio,bridge=vmbr0,firewall=1"},
		Scsi:         vm.Scsi{Scsi0: fmt.Sprintf("file=%s:8", storageName)},
		ScsiHw:       vm.VirtioScsiPci,
		SearchDomain: network.SearchDomain,
		Serial:       vm.Serial{Serial0: "socket"},
		VGA:          "serial0",
	}
	return vmoptions
}

// URL encodes the ssh keys
func sshKeyUrlEncode(keys string) (encodedKeys string) {
	encodedKeys = url.PathEscape(keys + "\n")
	encodedKeys = strings.Replace(encodedKeys, "+", "%2B", -1)
	encodedKeys = strings.Replace(encodedKeys, "@", "%40", -1)
	encodedKeys = strings.Replace(encodedKeys, "=", "%3D", -1)
	return
}

func EnsureRunning(instance vm.VirtualMachine) error {
	// ensure instance is running
	switch instance.Status {
	case vm.ProcessStatusRunning:
		return nil
	case vm.ProcessStatusStopped:
		if err := instance.Start(vm.StartOption{}); err != nil {
			return err
		}
	case vm.ProcessStatusPaused:
		if err := instance.Resume(vm.ResumeOption{}); err != nil {
			return err
		}
	default:
		errors.Errorf("unexpected status : %s", instance.Status)
	}
	return nil
}

// func EnsureStopped(instance vm.VirtualMachine) error {
// 	return nil
// }
