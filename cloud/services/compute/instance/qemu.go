package instance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
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

	nodeName := s.scope.NodeName()
	vmid := s.scope.GetVMID()
	qemu, err := s.getQEMU(ctx, vmid)
	if err == nil { // if qemu is found, return it
		return qemu, nil
	}
	if !rest.IsNotFound(err) {
		log.Error(err, fmt.Sprintf("failed to get qemu: node=%s,vmid=%d", nodeName, *vmid))
		return nil, err
	}

	// no qemu found, create new one
	return s.createQEMU(ctx, nodeName, vmid)
}

// get QEMU gets proxmox vm from vmid
func (s *Service) getQEMU(ctx context.Context, vmid *int) (*proxmox.VirtualMachine, error) {
	if vmid != nil {
		return s.client.VirtualMachine(ctx, *vmid)
	}
	return nil, rest.NotFoundErr
}

func (s *Service) createQEMU(ctx context.Context, nodeName string, vmid *int) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)

	// get node
	if nodeName == "" {
		// temp solution
		node, err := s.getRandomNode(ctx)
		if err != nil {
			log.Error(err, "failed to get random node")
			return nil, err
		}
		nodeName = node.Node
		s.scope.SetNodeName(nodeName)
	}

	// if vmid is empty, generate new vmid
	if vmid == nil {
		nextid, err := s.getNextID(ctx)
		if err != nil {
			log.Error(err, "failed to get available vmid")
			return nil, err
		}
		vmid = &nextid
	}

	vmoption := s.generateVMOptions()
	vm, err := s.client.CreateVirtualMachine(ctx, nodeName, *vmid, vmoption)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to create qemu instance %s", vm.VM.Name))
		return nil, err
	}
	s.scope.SetVMID(*vmid)
	if err := s.scope.PatchObject(); err != nil {
		return nil, err
	}
	return vm, nil
}

func (s *Service) getNextID(ctx context.Context) (int, error) {
	return s.client.RESTClient().GetNextID(ctx)
}

func (s *Service) getNodes(ctx context.Context) ([]*api.Node, error) {
	return s.client.Nodes(ctx)
}

// GetRandomNode returns a node chosen randomly
func (s *Service) getRandomNode(ctx context.Context) (*api.Node, error) {
	nodes, err := s.getNodes(ctx)
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

func (s *Service) generateVMOptions() api.VirtualMachineCreateOptions {
	vmName := s.scope.Name()
	snippetStorageName := s.scope.GetStorage().SnippetStorage.Name
	imageStorageName := s.scope.GetStorage().ImageStorage.Name
	network := s.scope.GetNetwork()
	hardware := s.scope.GetHardware()
	options := s.scope.GetOptions()
	cicustom := fmt.Sprintf("user=%s:%s", snippetStorageName, userSnippetPath(vmName))
	ide2 := fmt.Sprintf("file=%s:cloudinit,media=cdrom", imageStorageName)
	scsi0 := fmt.Sprintf("%s:0,import-from=%s", imageStorageName, rawImageFilePath(s.scope.GetImage()))
	net0 := "model=virtio,bridge=vmbr0,firewall=1"

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
		OnBoot:        boolToInt8(options.OnBoot),
		OSType:        api.OSType(options.OSType),
		Protection:    boolToInt8(options.Protection),
		Reboot:        int(boolToInt8(options.Reboot)),
		Scsi:          api.Scsi{Scsi0: scsi0},
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
