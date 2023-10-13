package v1beta1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"

	"github.com/sp-yduck/proxmox-go/api"
)

// ServerRef is used for configuring Proxmox client
type ServerRef struct {
	// endpoint is the address of the Proxmox-VE REST API endpoint.
	Endpoint string `json:"endpoint"`

	// to do : client options like insecure tls verify

	// SecretRef is a reference for secret which contains proxmox login secrets
	SecretRef *ObjectReference `json:"secretRef"`
}

// ObjectReference is a reference to another Kubernetes object instance.
type ObjectReference struct {
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	Namespace string `json:"namespace,omitempty"`

	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// Image is the image to be provisioned
type Image struct {
	// +kubebuilder:validation:Pattern:=.*\.(iso|img|qcow2|qed|raw|vdi|vpc|vmdk)$
	// URL is a location of an image to deploy.
	// supported formats are iso/qcow2/qed/raw/vdi/vpc/vmdk.
	URL string `json:"url"`

	// Checksum
	// Always better to specify checksum otherwise cappx will download
	// same image for every time. If checksum is specified, cappx will try
	// to avoid downloading existing image.
	Checksum string `json:"checksum,omitempty"`

	// +kubebuilder:validation:Enum:=sha256;sha256sum;md5;md5sum
	// ChecksumType
	ChecksumType *string `json:"checksumType,omitempty"`
}

// Hardware
type Hardware struct {
	// amount of RAM for the VM in MiB : 16 ~
	// +kubebuilder:validation:Minimum:=16
	// +kubebuilder:default:=4096
	Memory int `json:"memory,omitempty"`

	// number of CPU cores : 1 ~
	// +kubebuilder:validation:Mimimum:=1
	// +kubebuilder:default:=2
	CPU int `json:"cpu,omitempty"`

	// emulated cpu type
	// CPUType string `json:"cpuType,omitempty"`

	// +kubebuilder:validation:Minimum:=1
	// The number of CPU sockets. Defaults to 1.
	Sockets int `json:"sockets,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	// Limit of CPU usage. If the computer has 2 CPUs, it has total of '2' CPU time.
	// Value '0' indicates no CPU limit. Defaults to 0.
	CPULimit int `json:"cpuLimit,omitempty"`

	// Select BIOS implementation. Defaults to seabios. seabios or ovmf.
	// Defaults to seabios.
	BIOS BIOS `json:"bios,omitempty"`

	// Specifies the QEMU machine type.
	// regex: (pc|pc(-i440fx)?-\d+(\.\d+)+(\+pve\d+)?(\.pxe)?|q35|pc-q35-\d+(\.\d+)+(\+pve\d+)?(\.pxe)?|virt(?:-\d+(\.\d+)+)?(\+pve\d+)?)
	// Machine string `json:"machine,omitempty"`

	// SCSI controller model
	// SCSIHardWare SCSIHardWare `json:"scsiHardWare,omitempty"`

	// boot disk size
	// +kubebuilder:validation:Pattern:=\+?\d+(\.\d+)?[KMGT]?
	// +kubebuilder:default:="50G"
	Disk string `json:"disk,omitempty"`

	// Storage name for the boot disk. If none is provided, the ProxmoxCluster storage name will be used
	// +optional
	StorageName string `json:"storage,omitempty"`
}

// Network
// cloud-init network configuration is configured through Proxmox API
// it may be migrated to raw yaml way from Proxmox API way in the future
type Network struct {
	// to do : should accept multiple IPConfig
	IPConfig IPConfig `json:"ipConfig,omitempty"`

	// DNS server
	NameServer string `json:"nameServer,omitempty"`

	// search domain
	SearchDomain string `json:"searchDomain,omitempty"`

	// +kubebuilder:default:="virtio"
	Model string `json:"model,omitempty"`

	// +kubebuilder:default:="vmbr0"
	Bridge string `json:"bridge,omitempty"`

	Tag int `json:"vlanTag,omitempty"`
}

// IPConfig defines IP addresses and gateways for corresponding interface.
// it defaults to using dhcp on IPv4 if neither IP nor IP6 is specified.
type IPConfig struct {
	// IPv4 with CIDR
	IP string `json:"ip,omitempty"`

	// gateway IPv4
	Gateway string `json:"gateway,omitempty"`

	// IPv6 with CIDR
	IP6 string `json:"ip6,omitempty"`

	// gateway IPv6
	Gateway6 string `json:"gateway6,omitempty"`

	// IPv4FromPoolRef is a reference to an IP pool to allocate an address from.
	IPv4FromPoolRef *corev1.TypedLocalObjectReference `json:"IPv4FromPoolRef,omitempty"`

	// IPv6FromPoolRef is a reference to an IP pool to allocate an address from.
	IPv6FromPoolRef *corev1.TypedLocalObjectReference `json:"IPv6FromPoolRef,omitempty"`
}

func (c *IPConfig) String() string {
	configs := []string{}
	if c.IP != "" {
		configs = append(configs, fmt.Sprintf("ip=%s", c.IP))
	}
	if c.Gateway != "" {
		configs = append(configs, fmt.Sprintf("gw=%s", c.Gateway))
	}
	if c.IP6 != "" {
		configs = append(configs, fmt.Sprintf("ip6=%s", c.IP6))
	}
	if c.Gateway6 != "" {
		configs = append(configs, fmt.Sprintf("gw6=%s", c.Gateway6))
	}
	ipconfig := strings.Join(configs, ",")

	// it defaults to using dhcp on IPv4 if neither IP nor IP6 is specified
	if !strings.Contains(ipconfig, "ip") {
		ipconfig = "ip=dhcp"
	}

	return ipconfig
}

// Storage for image and snippets
type Storage struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type InstanceStatus string

var (
	InstanceStatusPaused  = InstanceStatus(api.ProcessStatusPaused)
	InstanceStatusRunning = InstanceStatus(api.ProcessStatusRunning)
	InstanceStatusStopped = InstanceStatus(api.ProcessStatusStopped)
)

type ClusterFailureDomainConfig struct {
	// Treat each node as a failure domain for cluster api
	// +optional
	NodeAsFailureDomain bool `json:"nodeAsFailureDomain,omitempty"`
}
