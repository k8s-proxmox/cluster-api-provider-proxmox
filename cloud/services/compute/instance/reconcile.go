package instance

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/providerid"
)

const (
	etcCAPPX = "/etc/cappx"
)

// reconcile normal
func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling instance")
	instance, err := s.createOrGetInstance(ctx)
	if err != nil {
		log.Error(err, "failed to create/get instance")
		return err
	}

	uuid, err := getBiosUUID(instance)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Reconciled instance: bios-uuid=%s", *uuid))
	if err := s.scope.SetProviderID(*uuid); err != nil {
		return err
	}
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))
	// s.scope.SetAddresses()
	return nil
}

// reconcile delete
func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleting instance resources")

	instance, err := s.GetInstance(ctx)
	if err != nil {
		if !IsNotFound(err) {
			return err
		}
		return nil
	}

	// must stop or pause instance before deletion
	// otherwise deletion will be fail
	if err := EnsureStoppedOrPaused(*instance); err != nil {
		return err
	}

	// delete cloud-config file
	if err := s.deleteCloudConfig(); err != nil {
		return err
	}

	// delete qemu
	return instance.Delete()
}

func (s *Service) createOrGetInstance(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	log.Info("Getting bootstrap data for machine")
	bootstrapData, err := s.scope.GetBootstrapData()
	if err != nil {
		log.Error(err, "Error getting bootstrap data for machine")
		return nil, errors.Wrap(err, "failed to retrieve bootstrap data")
	}

	if s.scope.GetBiosUUID() == nil {
		log.Info("ProxmoxMachine doesn't have bios UUID. instance will be created")
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
	biosUUID := s.scope.GetBiosUUID()
	if biosUUID == nil {
		return nil, api.ErrNotFound
	}
	vm, err := s.getInstanceFromBiosUUID(*biosUUID)
	if err != nil {
		if api.IsNotFound(err) {
			log.Info("instance wasn't found")
			return nil, api.ErrNotFound
		}
		log.Error(err, "failed to get instance from bios UUID")
		return nil, err
	}
	return vm, nil
}

func getBiosUUID(vm *vm.VirtualMachine) (*string, error) {
	config, err := vm.Config()
	if err != nil {
		return nil, err
	}
	smbios := config.SMBios1
	uuid, err := convertSMBiosToUUID(smbios)
	if err != nil {
		return nil, err
	}
	return pointer.String(uuid), nil
}

func convertSMBiosToUUID(smbios string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf("uuid=%s", providerid.UUIDFormat))
	match := re.FindString(smbios)
	if match == "" {
		return "", errors.Errorf("failed to fetch uuid form smbios")
	}
	// match: uuid=<uuid>
	return strings.Split(match, "=")[1], nil
}

func (s *Service) getInstanceFromBiosUUID(uuid string) (*vm.VirtualMachine, error) {
	nodes, err := s.client.Nodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("proxmox nodes not found")
	}

	// to do : check each node in parallel
	for _, node := range nodes {
		vms, err := node.VirtualMachines()
		if err != nil {
			continue
		}
		for _, vm := range vms {
			config, err := vm.Config()
			if err != nil {
				return nil, err
			}
			vmuuid, err := convertSMBiosToUUID(config.SMBios1)
			if err != nil {
				return nil, err
			}
			if vmuuid == uuid {
				return vm, nil
			}
		}
	}
	return nil, api.ErrNotFound
}

func (s *Service) CreateInstance(ctx context.Context, bootstrap string) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// qemu
	vm, err := s.reconcileQEMU(ctx)
	if err != nil {
		return nil, err
	}
	vmid := vm.VMID
	log.Info(fmt.Sprintf("reconciled qemu: node=%s,vmid=%d", vm.Node.Name(), vmid))

	// cloud init
	if err := s.reconcileCloudInit(bootstrap); err != nil {
		return nil, err
	}

	// set cloud image to hard disk and then resize
	if err := s.reconcileBootDevice(ctx, vm); err != nil {
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
		return errors.Errorf("unexpected status : %s", instance.Status)
	}
	return nil
}

func EnsureStoppedOrPaused(instance vm.VirtualMachine) error {
	switch instance.Status {
	case vm.ProcessStatusRunning:
		if err := instance.Stop(); err != nil {
			return err
		}
	case vm.ProcessStatusPaused:
		return nil
	case vm.ProcessStatusStopped:
		return nil
	default:
		return errors.Errorf("unexpected status : %s", instance.Status)
	}
	return nil
}
