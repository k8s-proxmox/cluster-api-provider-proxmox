package instance

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

func (s *Service) reconcileQEMU(ctx context.Context) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling QEMU")

	nodeName := s.scope.NodeName()
	vmid := s.scope.GetVMID()
	vm, err := s.getQEMU(nodeName, vmid)
	if err == nil { // if vm is found, return it
		return vm, nil
	}
	if !IsNotFound(err) {
		log.Error(err, fmt.Sprintf("failed to get vm: node=%s,vmid=%d", nodeName, *vmid))
		return nil, err
	}

	// no vm found, create new one
	return s.createQEMU(ctx, nodeName, vmid)
}

func (s *Service) getQEMU(nodeName string, vmid *int) (*vm.VirtualMachine, error) {
	if vmid != nil && nodeName != "" {
		node, err := s.GetNode(nodeName)
		if err != nil {
			return nil, err
		}
		return node.VirtualMachine(*vmid)
	}
	return nil, api.ErrNotFound
}

func (s *Service) createQEMU(ctx context.Context, nodeName string, vmid *int) (*vm.VirtualMachine, error) {
	log := log.FromContext(ctx)

	var node *node.Node
	var err error

	// get node
	if nodeName != "" {
		node, err = s.GetNode(nodeName)
		if err != nil {
			log.Error(err, fmt.Sprintf("failed to get node %s", nodeName))
			return nil, err
		}
	} else {
		// temp solution
		node, err = s.GetRandomNode()
		if err != nil {
			log.Error(err, "failed to get random node")
			return nil, err
		}
		s.scope.SetNodeName(node.Node)
	}

	// (for multiple node proxmox cluster support)
	// to do : set ssh client for specific node

	// if vmid is empty, generate new vmid
	if vmid == nil {
		nextid, err := s.GetNextID()
		if err != nil {
			log.Error(err, "failed to get available vmid")
			return nil, err
		}
		vmid = &nextid
	}

	// create vm
	vmoption := generateVMOptions(s.scope.Name(), s.scope.GetStorage().Name, s.scope.GetNetwork(), s.scope.GetHardware())
	vm, err := node.CreateVirtualMachine(*vmid, vmoption)
	if err != nil {
		log.Error(err, "failed to create virtual machine")
		return nil, err
	}
	s.scope.SetVMID(*vmid)
	return vm, nil
}

func (s *Service) GetNextID() (int, error) {
	return s.client.NextID()
}

func (s *Service) GetNodes() ([]*node.Node, error) {
	return s.client.Nodes()
}

func (s *Service) GetNode(name string) (*node.Node, error) {
	return s.client.Node(name)
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

func generateVMOptions(vmName, storageName string, network infrav1.Network, hardware infrav1.Hardware) vm.VirtualMachineCreateOptions {
	vmoptions := vm.VirtualMachineCreateOptions{
		Agent:        "enabled=1",
		Cores:        hardware.CPU,
		Memory:       hardware.Memory,
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
