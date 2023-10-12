package cloud

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

type Reconciler interface {
	Reconcile(ctx context.Context) error
	Delete(ctx context.Context) error
}

type Client interface {
	CloudClient() *proxmox.Service
	K8sClient() *client.Client
}

type Cluster interface {
	ClusterGetter
	ClusterSettter
}

// ClusterGetter is an interface which can get cluster information.
type ClusterGetter interface {
	Client
	Name() string
	Namespace() string
	// NetworkName() string
	// Network() *infrav1.Network
	// AdditionalLabels() infrav1.Labels
	FailureDomains() clusterv1.FailureDomains
	ControlPlaneEndpoint() clusterv1.APIEndpoint
	Storage() infrav1.Storage
}

type ClusterSettter interface {
	SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint)
	SetStorage(storage infrav1.Storage)
}

// MachineGetter is an interface which can get machine information.
type MachineGetter interface {
	Client
	GetProxmoxMachine() *infrav1.ProxmoxMachine
	Name() string
	Namespace() string
	// Zone() string
	// Role() string
	// IsControlPlane() bool
	// ControlPlaneGroupName() string
	NodeName() *string
	GetPool() *string
	GetBiosUUID() *string
	GetImage() infrav1.Image
	GetProviderID() string
	GetBootstrapData() (string, error)
	GetInstanceStatus() *infrav1.InstanceStatus
	GetClusterStorage() infrav1.Storage
	GetCloudInit() infrav1.CloudInit
	GetNetwork() infrav1.Network
	GetHardware() infrav1.Hardware
	GetBootDiskStorage() string
	GetVMID() *int
	GetOptions() infrav1.Options

	GetProxmoxMachineTemplate(context.Context) *infrav1.ProxmoxMachineTemplate
	GetProxmoxCluster() *infrav1.ProxmoxCluster
}

// MachineSetter is an interface which can set machine information.
type MachineSetter interface {
	SetProviderID(uuid string) error
	SetInstanceStatus(v infrav1.InstanceStatus)
	SetNodeName(name string)
	SetVMID(vmid int)
	SetPool(name string)
	SetConfigStatus(config api.VirtualMachineConfig)
	// SetFailureMessage(v error)
	// SetFailureReason(v capierrors.MachineStatusError)
	// SetAnnotation(key, value string)
	// SetAddresses(addressList []corev1.NodeAddress)
	SetFailureDomain(failureDomain string)
	PatchObject() error
}

// Machine is an interface which can get and set machine information.
type Machine interface {
	MachineGetter
	MachineSetter
}
