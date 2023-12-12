package instance

import (
	"context"
	"fmt"

	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/k8s-proxmox/proxmox-go/api"
	"github.com/k8s-proxmox/proxmox-go/proxmox"
	"github.com/k8s-proxmox/proxmox-go/rest"
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
	if err != nil {
		if !rest.IsNotFound(err) {
			log.Error(err, "failed to get qemu")
			return nil, err
		}

		// no qemu found, try to create new one
		log.V(3).Info("qemu wasn't found. new qemu will be created")
		if exist, err := s.client.VirtualMachineExistsWithName(ctx, s.scope.Name()); exist || err != nil {
			if exist {
				// there should no qemu with same name. occuring an error
				err = fmt.Errorf("qemu %s already exists", s.scope.Name())
			}
			log.Error(err, "stop creating new qemu to avoid replicating same qemu")
			return nil, err
		}
		qemu, err = s.createQEMU(ctx)
		if err != nil {
			log.Error(err, "failed to create qemu")
			return nil, err
		}
	}

	s.scope.SetVMID(qemu.VM.VMID)
	s.scope.SetNodeName(qemu.Node)
	if err := s.scope.PatchObject(); err != nil {
		return nil, err
	}
	return qemu, nil
}

// get QEMU gets proxmox vm from vmid
func (s *Service) getQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)
	log.Info("getting qemu from vmid")
	vmid := s.scope.GetVMID()
	if vmid != nil {
		return s.client.VirtualMachine(ctx, *vmid)
	}
	return nil, rest.NotFoundErr
}

func (s *Service) createQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)
	log.Info("creating qemu")

	if err := s.ensureStorageAvailable(ctx); err != nil {
		return nil, err
	}

	// create qemu
	log.Info("making qemu spec")
	spec := s.generateQemuSpec(ctx)
	result, err := s.scheduler.CreateQEMU(spec)
	if err != nil {
		log.Error(err, "failed to create qemu instance")
		return nil, err
	}
	return result.Instance(), nil
}

func (s *Service) generateVMOptions() api.VirtualMachineCreateOptions {
	vmName := s.scope.Name()
	snippetStorageName := s.scope.GetClusterStorage().Name
	imageStorageName := s.scope.GetStorage()
	network := s.scope.GetNetwork()
	hardware := s.scope.GetHardware()
	options := s.scope.GetOptions()
	cicustom := fmt.Sprintf("user=%s:%s", snippetStorageName, userSnippetPath(vmName))
	ide2 := fmt.Sprintf("file=%s:cloudinit,media=cdrom", imageStorageName)
	scsi0 := fmt.Sprintf("%s:0,import-from=%s", imageStorageName, rawImageFilePath(s.scope.GetImage()))
	net0 := hardware.NetworkDevice.String()

	vmoptions := api.VirtualMachineCreateOptions{
		ACPI:          boolToInt8(options.ACPI),
		Agent:         "enabled=1",
		Arch:          api.Arch(options.Arch),
		Balloon:       options.Balloon,
		BIOS:          string(hardware.BIOS),
		Boot:          fmt.Sprintf("order=%s", bootDvice),
		CiCustom:      cicustom,
		Cores:         hardware.CPU,
		CpuLimit:      hardware.CPULimit,
		Description:   options.Description,
		HugePages:     options.HugePages.String(),
		Ide:           api.Ide{Ide2: ide2},
		IPConfig:      api.IPConfig{IPConfig0: network.IPConfig.String()},
		KeepHugePages: boolToInt8(options.KeepHugePages),
		KVM:           boolToInt8(options.KVM),
		LocalTime:     boolToInt8(options.LocalTime),
		Lock:          string(options.Lock),
		Memory:        hardware.Memory,
		Name:          vmName,
		NameServer:    network.NameServer,
		Net:           api.Net{Net0: net0},
		Numa:          boolToInt8(options.NUMA),
		Node:          s.scope.NodeName(),
		OnBoot:        boolToInt8(options.OnBoot),
		OSType:        api.OSType(options.OSType),
		// Pool: pool,
		Protection:   boolToInt8(options.Protection),
		Reboot:       int(boolToInt8(options.Reboot)),
		Scsi:         api.Scsi{Scsi0: scsi0},
		ScsiHw:       api.VirtioScsiPci,
		SearchDomain: network.SearchDomain,
		Serial:       api.Serial{Serial0: "socket"},
		Shares:       options.Shares,
		Sockets:      hardware.Sockets,
		Storage:      imageStorageName,
		Tablet:       boolToInt8(options.Tablet),
		Tags:         options.Tags.String(),
		TDF:          boolToInt8(options.TimeDriftFix),
		Template:     boolToInt8(options.Template),
		VCPUs:        options.VCPUs,
		VMGenID:      options.VMGenerationID,
		VMID:         s.scope.GetVMID(),
		VGA:          "serial0",
	}
	return vmoptions
}

type QEMUSpec struct {
	Ctx          context.Context
	CloneOption  api.VirtualMachineCloneOption
	config       api.VirtualMachineConfig
	CreateOption api.VirtualMachineCreateOptions
}

func (s *Service) generateQemuSpec(ctx context.Context) *QEMUSpec {
	// bind annotation key-values to context
	schedCtx := framework.ContextWithMap(ctx, s.scope.Annotations())
	createOption := s.generateVMOptions()
	clone := s.scope.GetCloneSpec()
	cloneOption := api.VirtualMachineCloneOption{}
	config := createOptionToConfig(createOption)
	if clone.TemplateID != nil {
		cloneOption = api.VirtualMachineCloneOption{
			VMID:    *clone.TemplateID,
			BWLimit: createOption.BWLimit,
			Full:    boolToInt8(true),
			Name:    s.scope.Name(),
			Pool:    createOption.Pool,
			Storage: createOption.Storage,
		}
		cloneOption.Description += fmt.Sprintf("\n\n*Cloned from %d", *clone.TemplateID)
	}
	return &QEMUSpec{Ctx: schedCtx, CloneOption: cloneOption, config: config, CreateOption: createOption}
}

func (q *QEMUSpec) Name() string {
	return q.CreateOption.Name
}

func (q *QEMUSpec) Context() context.Context {
	return q.Ctx
}

func (q *QEMUSpec) CloneSpec() *api.VirtualMachineCloneOption {
	return &q.CloneOption
}

func (q *QEMUSpec) Config() *api.VirtualMachineConfig {
	return &q.config
}

func (q *QEMUSpec) CreateSpec() *api.VirtualMachineCreateOptions {
	return &q.CreateOption
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

func createOptionToConfig(opts api.VirtualMachineCreateOptions) api.VirtualMachineConfig {
	return api.VirtualMachineConfig{
		ACPI:          opts.ACPI,
		Agent:         opts.Agent,
		Arch:          opts.Arch,
		Balloon:       opts.Balloon,
		BIOS:          opts.BIOS,
		Boot:          opts.Boot,
		CiCustom:      opts.CiCustom,
		Cores:         opts.Cores,
		CpuLimit:      opts.CpuLimit,
		Description:   opts.Description,
		HugePages:     opts.HugePages,
		Ide:           opts.Ide,
		IPConfig:      opts.IPConfig,
		KeepHugePages: opts.KeepHugePages,
		KVM:           opts.KVM,
		LocalTime:     opts.LocalTime,
		Lock:          opts.Lock,
		Memory:        opts.Memory,
		Name:          opts.Name,
		NameServer:    opts.NameServer,
		Net:           opts.Net,
		Numa:          opts.Numa,
		Node:          opts.Node,
		OnBoot:        opts.OnBoot,
		OSType:        opts.OSType,
		Protection:    opts.Protection,
		Reboot:        opts.Reboot,
		Scsi:          opts.Scsi,
		ScsiHw:        opts.ScsiHw,
		SearchDomain:  opts.SearchDomain,
		Serial:        opts.Serial,
		Shares:        opts.Shares,
		Sockets:       opts.Sockets,
		Tablet:        opts.Tablet,
		Tags:          opts.Tags,
		TDF:           opts.TDF,
		Template:      opts.Template,
		VCPUs:         opts.VCPUs,
		VMGenID:       opts.VMGenID,
		VGA:           opts.VGA,
	}
}
