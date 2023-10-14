package cloud

import (
	"context"

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
	// FailureDomains() clusterv1.FailureDomains
	ControlPlaneEndpoint() clusterv1.APIEndpoint
}

type ClusterSettter interface {
	SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint)
}

// MachineGetter is an interface which can get machine information.
type MachineGetter interface {
	Client
	ClusterName() string
	Name() string
	Namespace() string
	// Zone() string
	// Role() string
	// IsControlPlane() bool
	// ControlPlaneGroupName() string
	NodeName() string
	GetBiosUUID() *string
	GetImage() infrav1.Image
	GetProviderID() string
	GetBootstrapData() (string, error)
	GetInstanceStatus() *infrav1.InstanceStatus
	GetStorage() infrav1.Storage
	GetCloudInit() infrav1.CloudInit
	GetNetwork() infrav1.Network
	GetHardware() infrav1.Hardware
	GetVMID() *int
	GetOptions() infrav1.Options
}

// MachineSetter is an interface which can set machine information.
type MachineSetter interface {
	SetProviderID(uuid string) error
	SetInstanceStatus(status infrav1.InstanceStatus)
	SetNodeName(name string)
	SetVMID(vmid int)
	SetSnippetStorage(storage infrav1.SnippetStorage)
	SetImageStorage(storage infrav1.ImageStorage)
	SetConfigStatus(config api.VirtualMachineConfig)
	// SetFailureMessage(v error)
	// SetFailureReason(v capierrors.MachineStatusError)
	// SetAnnotation(key, value string)
	// SetAddresses(addressList []corev1.NodeAddress)
	PatchObject() error
}

// Machine is an interface which can get and set machine information.
type Machine interface {
	MachineGetter
	MachineSetter
}
