package cloud

import (
	"context"

	"github.com/sp-yduck/proxmox/pkg/service"
	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scope"
)

type Reconciler interface {
	Reconcile(ctx context.Context) error
	Delete(ctx context.Context) error
}

type Client interface {
	CloudClient() *service.Service
	RemoteClient() *scope.SSHClient
}

type Cluster interface {
	ClusterGetter
	ClusterSettter
}

// ClusterGetter is an interface which can get cluster information.
type ClusterGetter interface {
	Client
	// Region() string
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
	Name() string
	Namespace() string
	// Zone() string
	// Role() string
	// IsControlPlane() bool
	// ControlPlaneGroupName() string
	GetInstanceID() *string
	GetProviderID() string
	GetBootstrapData() (string, error)
	GetInstanceStatus() *infrav1.InstanceStatus
	GetStorage() infrav1.Storage
	GetCloudInit() infrav1.CloudInit
	GetNetwork() infrav1.Network
}

// MachineSetter is an interface which can set machine information.
type MachineSetter interface {
	SetProviderID(instance *vm.VirtualMachine) error
	SetInstanceStatus(v infrav1.InstanceStatus)
	// SetFailureMessage(v error)
	// SetFailureReason(v capierrors.MachineStatusError)
	// SetAnnotation(key, value string)
	// SetAddresses(addressList []corev1.NodeAddress)
}

// Machine is an interface which can get and set machine information.
type Machine interface {
	MachineGetter
	MachineSetter
}
