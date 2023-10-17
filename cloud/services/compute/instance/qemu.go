package instance

import (
	"context"
	"fmt"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"github.com/sp-yduck/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	bootDvice = "scsi0"
)

// reconciles QEMU instance
func (s *Service) reconcileQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling QEMU")

	qemu, err := s.getQEMU(ctx)
	if err == nil { // if qemu is found, return it
		return qemu, nil
	}
	if !rest.IsNotFound(err) {
		log.Error(err, "failed to get qemu")
		return nil, err
	}

	// no qemu found, create new one
	return s.createQEMU(ctx)
}

// get QEMU gets proxmox vm from vmid
func (s *Service) getQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	vmid := s.scope.GetVMID()
	if vmid != nil {
		return s.client.VirtualMachine(ctx, *vmid)
	}
	return nil, rest.NotFoundErr
}

func (s *Service) createQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// bind annotation key-values to context
	schedCtx := framework.ContextWithMap(ctx, s.scope.Annotations())

	// node assignment
	vmoption := s.generateVMOptions()
	nodeName, err := s.scheduler.SelectNode(schedCtx, vmoption)
	if err != nil {
		log.Error(err, "failed to select proxmox node")
		return nil, err
	}
	s.scope.SetNodeName(nodeName)

	// vmid assignment
	vmid, err := s.scheduler.SelectVMID(schedCtx, vmoption)
	if err != nil {
		log.Error(err, "failed to get available vmid")
		return nil, err
	}

	vm, err := s.client.CreateVirtualMachine(ctx, nodeName, vmid, vmoption)
	if err != nil {
		log.Error(err, "failed to create qemu instance")
		return nil, err
	}
	s.scope.SetVMID(vmid)
	if err := s.scope.PatchObject(); err != nil {
		return nil, err
	}
	return vm, nil
}

func (s *Service) generateVMOptions() api.VirtualMachineCreateOptions {
	vmName := s.scope.Name()
	storageName := s.scope.GetStorage().Name
	network := s.scope.GetNetwork()
	hardware := s.scope.GetHardware()
	options := s.scope.GetOptions()

	vmoptions := api.VirtualMachineCreateOptions{
		ACPI:          boolToInt8(options.ACPI),
		Agent:         "enabled=1",
		Arch:          api.Arch(options.Arch),
		Balloon:       options.Balloon,
		BIOS:          string(hardware.BIOS),
		Boot:          fmt.Sprintf("order=%s", bootDvice),
		CiCustom:      fmt.Sprintf("user=%s:%s", storageName, userSnippetPath(vmName)),
		Cores:         hardware.CPU,
		CpuLimit:      hardware.CPULimit,
		Description:   options.Description,
		HugePages:     options.HugePages.String(),
		Ide:           api.Ide{Ide2: fmt.Sprintf("file=%s:cloudinit,media=cdrom", storageName)},
		IPConfig:      api.IPConfig{IPConfig0: network.IPConfig.String()},
		KeepHugePages: boolToInt8(options.KeepHugePages),
		KVM:           boolToInt8(options.KVM),
		LocalTime:     boolToInt8(options.LocalTime),
		Lock:          string(options.Lock),
		Memory:        hardware.Memory,
		Name:          vmName,
		NameServer:    network.NameServer,
		Net:           api.Net{Net0: "model=virtio,bridge=vmbr0,firewall=1"},
		Numa:          boolToInt8(options.NUMA),
		Node:          s.scope.NodeName(),
		OnBoot:        boolToInt8(options.OnBoot),
		OSType:        api.OSType(options.OSType),
		Protection:    boolToInt8(options.Protection),
		Reboot:        int(boolToInt8(options.Reboot)),
		Scsi:          api.Scsi{Scsi0: fmt.Sprintf("file=%s:8", storageName)},
		ScsiHw:        api.VirtioScsiPci,
		SearchDomain:  network.SearchDomain,
		Serial:        api.Serial{Serial0: "socket"},
		Shares:        options.Shares,
		Sockets:       hardware.Sockets,
		Tablet:        boolToInt8(options.Tablet),
		Tags:          options.Tags.String(),
		TDF:           boolToInt8(options.TimeDriftFix),
		Template:      boolToInt8(options.Template),
		VCPUs:         options.VCPUs,
		VMGenID:       options.VMGenerationID,
		VMID:          s.scope.GetVMID(),
		VGA:           "serial0",
	}
	return vmoptions
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
