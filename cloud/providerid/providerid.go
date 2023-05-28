package providerid

import (
	"fmt"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const Prefix = "proxmox://"

type ProviderID interface {
	Node() string
	VMID() int
	fmt.Stringer
}

type providerID struct {
	// proxmox node name
	node string
	// proxmox vmid
	vmid int
}

func New(node string, vmid int) (ProviderID, error) {
	if node == "" {
		return nil, errors.New("location required for provider id")
	}
	if vmid == 0 {
		return nil, errors.New("vmid required for provider id")
	}

	return &providerID{
		node: node,
		vmid: vmid,
	}, nil
}

func (p *providerID) Node() string {
	return p.node
}

func (p *providerID) VMID() int {
	return p.vmid
}

func (p *providerID) String() string {
	return Prefix + path.Join(p.node, strconv.Itoa(p.vmid))
}
