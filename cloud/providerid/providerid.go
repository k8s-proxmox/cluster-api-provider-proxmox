package providerid

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
)

const Prefix = "proxmox://"

type ProviderID interface {
	Cluster() string
	Node() string
	Name() string
	fmt.Stringer
}

type providerID struct {
	// proxmox cluster name
	cluster string
	// proxmox node name
	node string
	// proxmox vm name
	name string
}

func New(cluster, node, name string) (ProviderID, error) {
	if cluster == "" {
		return nil, errors.New("project required for provider id")
	}
	if node == "" {
		return nil, errors.New("location required for provider id")
	}
	if name == "" {
		return nil, errors.New("name required for provider id")
	}

	return &providerID{
		cluster: cluster,
		node:    node,
		name:    name,
	}, nil
}

func (p *providerID) Cluster() string {
	return p.cluster
}

func (p *providerID) Node() string {
	return p.node
}

func (p *providerID) Name() string {
	return p.name
}

func (p *providerID) String() string {
	return Prefix + path.Join(p.cluster, p.node, p.name)
}
