package instance

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"github.com/sp-yduck/proxmox-go/rest"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
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

	uuid, err := getBiosUUIDFromVM(ctx, instance)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Reconciled instance: bios-uuid=%s", *uuid))
	if err := s.scope.SetProviderID(*uuid); err != nil {
		return err
	}
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.VM.Status))
	s.scope.SetNodeName(instance.Node)
	s.scope.SetVMID(instance.VM.VMID)

	config, err := instance.GetConfig(ctx)
	if err != nil {
		return err
	}
	s.scope.SetConfigStatus(*config)
	return nil
}

// reconcile delete
func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleting instance resources")

	instance, err := s.getQEMU(ctx, s.scope.GetVMID())
	if err != nil {
		if !rest.IsNotFound(err) {
			return err
		}
		log.Info("qemu is not found or already deleted")
		return nil
	}

	// must stop or pause instance before deletion
	// otherwise deletion will be fail
	if err := ensureStoppedOrPaused(ctx, *instance); err != nil {
		return err
	}

	// delete cloud-config file
	if err := s.deleteCloudConfig(ctx); err != nil {
		return err
	}

	// delete qemu
	return instance.Delete(ctx)
}

func (s *Service) createOrGetInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)

	instance, err := s.getInstance(ctx)
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("instance wasn't found. new instance will be created")
			return s.createInstance(ctx)
		}
		log.Error(err, "failed to get instance")
		return nil, err
	}

	return instance, nil
}

// getInstance() gets proxmoxm vm from providerID
func (s *Service) getInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)

	biosUUID := s.scope.GetBiosUUID()
	if biosUUID == nil {
		log.Info("instance does not have providerID yet")
		return nil, rest.NotFoundErr
	}

	vm, err := s.client.VirtualMachineFromUUID(ctx, *biosUUID)
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("instance wasn't found")
			return nil, rest.NotFoundErr
		}
		log.Error(err, "failed to get instance from bios UUID")
		return nil, err
	}

	return vm, nil
}

func getBiosUUIDFromVM(ctx context.Context, vm *proxmox.VirtualMachine) (*string, error) {
	log := log.FromContext(ctx)
	config, err := vm.GetConfig(ctx)
	if err != nil {
		log.Error(err, "failed to get vm config")
		return nil, err
	}
	smbios := config.SMBios1
	uuid, err := proxmox.ConvertSMBiosToUUID(smbios)
	if err != nil {
		log.Error(err, "failed to convert SMBios to UUID")
		return nil, err
	}
	return pointer.String(uuid), nil
}

func (s *Service) createInstance(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// qemu
	instance, err := s.reconcileQEMU(ctx)
	if err != nil {
		return nil, err
	}
	vmid := instance.VM.VMID
	log.Info(fmt.Sprintf("reconciled qemu: node=%s,vmid=%d", instance.Node, vmid))

	// cloud init
	if err := s.reconcileCloudInit(ctx); err != nil {
		return nil, err
	}

	// set cloud image to hard disk and then resize
	if err := s.reconcileBootDevice(ctx, instance); err != nil {
		return nil, err
	}

	// vm status
	if err := ensureRunning(ctx, *instance); err != nil {
		return nil, err
	}
	return instance, nil
}

func ensureRunning(ctx context.Context, instance proxmox.VirtualMachine) error {
	log := log.FromContext(ctx)
	// ensure instance is running
	switch instance.VM.Status {
	case api.ProcessStatusRunning:
		return nil
	case api.ProcessStatusStopped:
		if err := instance.Start(ctx, api.VirtualMachineStartOption{}); err != nil {
			log.Error(err, "failed to start instance process")
			return err
		}
	case api.ProcessStatusPaused:
		if err := instance.Resume(ctx, api.VirtualMachineResumeOption{}); err != nil {
			log.Error(err, "failed to resume instance process")
			return err
		}
	default:
		return errors.Errorf("unexpected status : %s", instance.VM.Status)
	}
	return nil
}

func ensureStoppedOrPaused(ctx context.Context, instance proxmox.VirtualMachine) error {
	log := log.FromContext(ctx)
	switch instance.VM.Status {
	case api.ProcessStatusRunning:
		if err := instance.Stop(ctx, api.VirtualMachineStopOption{}); err != nil {
			log.Error(err, "failed to stop instance process")
			return err
		}
	case api.ProcessStatusPaused, api.ProcessStatusStopped:
		return nil
	default:
		return errors.Errorf("unexpected status : %s", instance.VM.Status)
	}
	return nil
}
