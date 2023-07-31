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

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

const (
	bootDvice = "scsi0"
)

func (s *Service) reconcileQEMU(ctx context.Context) (*proxmox.VirtualMachine, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling QEMU")

	nodeName := s.scope.NodeName()
	vmid := s.scope.GetVMID()
	api, err := s.getQEMU(ctx, vmid)
	if err == nil { // if api is found, return it
		return api, nil
	}
	if !rest.IsNotFound(err) {
		log.Error(err, fmt.Sprintf("failed to get api: node=%s,vmid=%d", nodeName, *vmid))
		return nil, err
	}

	// no api found, create new one
	return s.createQEMU(ctx, nodeName, vmid)
}

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
		node, err := s.getRandomNode()
		if err != nil {
			log.Error(err, "failed to get random node")
			return nil, err
		}
		nodeName = node.Node
		s.scope.SetNodeName(nodeName)
	}

	// if vmid is empty, generate new vmid
	if vmid == nil {
		nextid, err := s.getNextID()
		if err != nil {
			log.Error(err, "failed to get available vmid")
			return nil, err
		}
		vmid = &nextid
	}

	vmoption := generateVMOptions(s.scope.Name(), s.scope.GetStorage().Name, s.scope.GetNetwork(), s.scope.GetHardware())
	vm, err := s.client.CreateVirtualMachine(ctx, nodeName, *vmid, vmoption)
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to create qemu instance %s", vm.VM.Name))
		return nil, err
	}
	s.scope.SetVMID(*vmid)
	return vm, nil
}

func (s *Service) getNextID() (int, error) {
	return s.client.RESTClient().GetNextID(context.TODO())
}

func (s *Service) getNodes() ([]*api.Node, error) {
	return s.client.Nodes(context.TODO())
}

// GetRandomNode returns a node chosen randomly
func (s *Service) getRandomNode() (*api.Node, error) {
	nodes, err := s.getNodes()
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

func generateVMOptions(vmName, storageName string, network infrav1.Network, hardware infrav1.Hardware) api.VirtualMachineCreateOptions {
	vmoptions := api.VirtualMachineCreateOptions{
		Agent:        "enabled=1",
		Cores:        hardware.CPU,
		Memory:       hardware.Memory,
		Name:         vmName,
		NameServer:   network.NameServer,
		Boot:         fmt.Sprintf("order=%s", bootDvice),
		Ide:          api.Ide{Ide2: fmt.Sprintf("file=%s:cloudinit,media=cdrom", storageName)},
		CiCustom:     fmt.Sprintf("user=%s:%s", storageName, userSnippetPath(vmName)),
		IPConfig:     api.IPConfig{IPConfig0: network.IPConfig.String()},
		OSType:       api.L26,
		Net:          api.Net{Net0: "model=virtio,bridge=vmbr0,firewall=1"},
		Scsi:         api.Scsi{Scsi0: fmt.Sprintf("file=%s:8", storageName)},
		ScsiHw:       api.VirtioScsiPci,
		SearchDomain: network.SearchDomain,
		Serial:       api.Serial{Serial0: "socket"},
		VGA:          "serial0",
	}
	return vmoptions
}
