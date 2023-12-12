package cloud

import (
	"context"

	"github.com/k8s-proxmox/proxmox-go/api"
	"github.com/k8s-proxmox/proxmox-go/proxmox"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/k8s-proxmox/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/k8s-proxmox/cluster-api-provider-proxmox/cloud/scheduler"
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
	Storage() infrav1.Storage
}

type ClusterSettter interface {
	SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint)
	SetStorage(storage infrav1.Storage)
}

// MachineGetter is an interface which can get machine information.
type MachineGetter interface {
	Client
	GetScheduler(client *proxmox.Service) *scheduler.Scheduler
	Name() string
	Namespace() string
	Annotations() map[string]string
	// Zone() string
	// Role() string
	// IsControlPlane() bool
	// ControlPlaneGroupName() string
	NodeName() string
	GetBiosUUID() *string
	GetImage() infrav1.Image
	GetCloneSpec() infrav1.CloneSpec
	GetProviderID() string
	GetBootstrapData() (string, error)
	GetInstanceStatus() *infrav1.InstanceStatus
	GetClusterStorage() infrav1.Storage
	GetStorage() string
	GetCloudInit() infrav1.CloudInit
	GetNetwork() infrav1.Network
	GetHardware() infrav1.Hardware
	GetVMID() *int
	GetOptions() infrav1.Options
}

// MachineSetter is an interface which can set machine information.
type MachineSetter interface {
	SetProviderID(uuid string) error
	SetInstanceStatus(v infrav1.InstanceStatus)
	SetNodeName(name string)
	SetVMID(vmid int)
	SetConfigStatus(config api.VirtualMachineConfig)
	SetStorage(name string)
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
