package instance

import (
	"context"
	"fmt"
	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"math"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"github.com/sp-yduck/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	bootDevice = "scsi0"
)

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
		log.Error(err, fmt.Sprintf("failed to get qemu: node=%s,vmid=%d", *nodeName, *vmid))
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

func (s *Service) createQEMU(ctx context.Context, nodeName *string, vmid *int) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)
	template := s.scope.GetProxmoxMachineTemplate(ctx)
	cluster := s.scope.GetProxmoxCluster()

	if cluster.Spec.ResourcePool != "" {
		s.scope.SetPool(cluster.Spec.ResourcePool)
		if err := s.scope.PatchObject(); err != nil {
			return nil, err
		}
	}

	// get node
	if nodeName == nil {
		if template.Spec.Nodes != nil && len(template.Spec.Nodes) > 0 {
			log.Info("selecting random node from configured nodes in template")
			nodeName = &template.Spec.Nodes[rand.Intn(len(template.Spec.Nodes))]
		} else if cluster.Spec.Nodes != nil && len(cluster.Spec.Nodes) > 0 {
			log.Info("selecting random node from configured nodes in cluster")
			nodeName = &cluster.Spec.Nodes[rand.Intn(len(cluster.Spec.Nodes))]
		} else {
			log.Info("selecting random node")
			node, err := s.getRandomNode(ctx)
			if err != nil {
				log.Error(err, "failed to get random node")
				return nil, err
			}
			nodeName = &node.Node
		}
		s.scope.SetNodeName(*nodeName)
		if s.scope.GetProxmoxMachine().Spec.FailureDomain == nil && s.scope.GetProxmoxCluster().Spec.FailureDomainConfig != nil && s.scope.GetProxmoxCluster().Spec.FailureDomainConfig.NodeAsFailureDomain {
			s.scope.SetFailureDomain(*nodeName)
		}
		if err := s.scope.PatchObject(); err != nil {
			return nil, err
		}
	}

	// if vmid is empty, generate new vmid
	if vmid == nil {
		if template != nil && template.Spec.VMIDs != nil {
			nextid, err := s.getNextVmIdInConfiguredRange(ctx, template)
			if err != nil {
				return nil, err
			}
			vmid = &nextid
		} else {
			log.Info("using next id from proxmox cluster as VM ID")
			nextid, err := s.getNextID(ctx)
			if err != nil {
				log.Error(err, "failed to get available vmid")
				return nil, err
			}
			vmid = &nextid
		}
		s.scope.SetVMID(*vmid)
		if err := s.scope.PatchObject(); err != nil {
			return nil, err
		}
		log.Info(fmt.Sprintf("new vm id %d", *vmid))
	}

	// os image
	image, err := s.importCloudImage(ctx)
	if err != nil {
		return nil, err
	}

	config, err := s.generateIpConfig0(ctx)
	if err != nil {
		return nil, err
	}

	vmoption := s.generateVMOptions(image, config)
	vm, err := s.client.CreateVirtualMachine(ctx, *nodeName, *vmid, vmoption)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to create qemu instance %d", &vmid))
		return nil, err
	}
	return vm, nil
}

func (s *Service) getNextVmIdInConfiguredRange(ctx context.Context, template *infrav1.ProxmoxMachineTemplate) (int, error) {
	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("generating VM ID based on range in configured range %d-%d", template.Spec.VMIDs.Start, template.Spec.VMIDs.End))

	vms, err := s.client.VirtualMachines(ctx)
	if err != nil {
		log.Error(err, "failed to get virtual machines")
		return 0, err
	}

	usedVmIds := map[int]bool{}
	for _, vm := range vms {
		usedVmIds[vm.VMID] = true
	}

	nextid := template.Spec.VMIDs.Start

	var maxId int
	if template.Spec.VMIDs.End > 0 {
		maxId = template.Spec.VMIDs.End
	} else {
		maxId = math.MaxInt32 - 1
	}

	for nextid <= maxId {
		_, isUsed := usedVmIds[nextid]
		if !isUsed {
			break
		}
		nextid += 1
	}

	if nextid > maxId {
		log.Error(err, "no available VM ID found")
		return 0, err
	}
	return nextid, nil
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

func (s *Service) generateVMOptions(importFromImage string, ipconfig0 string) api.VirtualMachineCreateOptions {
	vmName := s.scope.Name()
	pool := s.scope.GetPool()
	clusterStorageName := s.scope.GetClusterStorage().Name
	network := s.scope.GetNetwork()
	hardware := s.scope.GetHardware()
	options := s.scope.GetOptions()

	if ipconfig0 == "" {
		ipconfig0 = network.IPConfig.String()
	}

	net0 := fmt.Sprintf("model=%s,bridge=%s,firewall=1", network.Model, network.Bridge)
	if network.Tag > 0 {
		net0 += fmt.Sprintf(",tag=%d", network.Tag)
	}

	cloudinitDiskConfig := fmt.Sprintf("file=%s:cloudinit,media=cdrom", s.scope.GetBootDiskStorage())
	bootDiskConfig := fmt.Sprintf("%s:0,import-from=%s", s.scope.GetBootDiskStorage(), importFromImage)

	vmoptions := api.VirtualMachineCreateOptions{
		ACPI:          boolToInt8(options.ACPI),
		Agent:         "enabled=1",
		Arch:          api.Arch(options.Arch),
		AutoStart:     boolToInt8(true),
		Balloon:       options.Balloon,
		BIOS:          string(hardware.BIOS),
		Boot:          fmt.Sprintf("order=%s", bootDevice),
		CiCustom:      fmt.Sprintf("user=%s:%s", clusterStorageName, userSnippetPath(vmName)),
		Cores:         hardware.CPU,
		CpuLimit:      hardware.CPULimit,
		Description:   options.Description,
		HugePages:     options.HugePages.String(),
		Ide:           api.Ide{Ide2: cloudinitDiskConfig},
		IPConfig:      api.IPConfig{IPConfig0: ipconfig0},
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
		Scsi:          api.Scsi{Scsi0: bootDiskConfig},
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
	if pool != nil {
		vmoptions.Pool = *pool
	}
	return vmoptions
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
