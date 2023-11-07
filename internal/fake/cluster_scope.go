package fake

import (
	"github.com/sp-yduck/proxmox-go/proxmox"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

type FakeClusterScope struct {
	cloudClient          *proxmox.Service
	name                 string
	namespace            string
	controlPlaneEndpoint clusterv1.APIEndpoint
	storage              infrav1.Storage
	resourcePool         string
}

func NewClusterScope(client *proxmox.Service) *FakeClusterScope {
	return &FakeClusterScope{
		cloudClient: client,
		name:        "foo-cluster",
		namespace:   "default",
		controlPlaneEndpoint: clusterv1.APIEndpoint{
			Host: "foo-host",
			Port: 6443,
		},
	}
}

func (f *FakeClusterScope) Name() string {
	return f.name
}

func (f *FakeClusterScope) Namespace() string {
	return f.namespace
}

func (f *FakeClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	return f.controlPlaneEndpoint
}

func (f *FakeClusterScope) Storage() infrav1.Storage {
	return f.storage
}

func (f *FakeClusterScope) ResourcePool() string {
	if f.resourcePool == "" {
		return f.name
	}
	return f.resourcePool
}

func (f *FakeClusterScope) CloudClient() *proxmox.Service {
	return f.cloudClient
}

func (f *FakeClusterScope) SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint) {
	f.controlPlaneEndpoint = endpoint
}

func (f *FakeClusterScope) SetStorage(storage infrav1.Storage) {
	f.storage = storage
}

func (f *FakeClusterScope) SetName(name string) {
	f.name = name
}
