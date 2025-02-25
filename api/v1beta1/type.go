package v1beta1

import (
	"fmt"
	"strings"

	"github.com/k8s-proxmox/proxmox-go/api"
)

type InstanceStatus string

var (
	InstanceStatusPaused  = InstanceStatus(api.ProcessStatusPaused)
	InstanceStatusRunning = InstanceStatus(api.ProcessStatusRunning)
	InstanceStatusStopped = InstanceStatus(api.ProcessStatusStopped)
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

// ExtraDisk represents an additional virtual disk
type ExtraDisk struct {
	Size    string `json:"size,omitempty"`    // e.g., "100Gi"
	Storage string `json:"storage,omitempty"` // e.g., "local-lvm"
	Type    string `json:"type,omitempty"`    // e.g., "scsi", "virtio"
	Format  string `json:"format,omitempty"`  //
}

// Hardware
type Hardware struct {
	// amount of RAM for the VM in MiB : 16 ~
	// +kubebuilder:validation:Minimum:=16
	// +kubebuilder:default:=4096
	Memory int `json:"memory,omitempty"`

	// number of CPU cores : 1 ~
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:default:=2
	CPU int `json:"cpu,omitempty"`

	// Emulated CPU Type. Defaults to kvm64
	CPUType string `json:"cpuType,omitempty"`

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

	// hard disk size
	// +kubebuilder:validation:Pattern:=\+?\d+(\.\d+)?[KMGT]?
	// +kubebuilder:default:="50G"
	RootDisk   string      `json:"rootDisk,omitempty"`
	ExtraDisks []ExtraDisk `json:"extraDisks,omitempty"` // Add support for multiple disks

	// network devices
	// to do: multiple devices
	// +kubebuilder:default:={model:virtio,bridge:vmbr0,firewall:true}
	NetworkDevice NetworkDevice `json:"networkDevice,omitempty"`
}

// Network Device
type NetworkDevice struct {
	// +kubebuilder:default:="virtio"
	Model NetworkDeviceModel `json:"model,omitempty"`

	// +kubebuilder:default:="vmbr0"
	Bridge NetworkDeviceBridge `json:"bridge,omitempty"`

	// +kubebuilder:default:=true
	Firewall bool `json:"firewall,omitempty"`

	LinkDown bool `json:"linkDown,omitempty"`

	MacAddr string `json:"macAddr,omitempty"`

	MTU int `json:"mtu,omitempty"`

	Queues int `json:"queues,omitempty"`

	// since float is highly discouraged, use string instead
	// +kubebuilder:validation:Pattern:=[0-9]+(\.|)[0-9]*
	Rate string `json:"rate,omitempty"`

	Tag int `json:"tag,omitempty"`

	// trunks: array of vlanid
	Trunks []int `json:"trunks,omitempty"`
}

type (
	// +kubebuilder:validation:Enum:=e1000;virtio;rtl8139;vmxnet3
	NetworkDeviceModel string

	// +kubebuilder:validation:Pattern:="vmbr[0-9]{1,4}"
	NetworkDeviceBridge string
)

func (n *NetworkDevice) String() string {
	config := []string{}
	config = append(config, fmt.Sprintf("model=%s", string(n.Model)))
	if n.Bridge != "" {
		config = append(config, fmt.Sprintf("bridge=%s", string(n.Bridge)))
	}
	if n.Firewall {
		config = append(config, fmt.Sprintf("firewall=%d", btoi(n.Firewall)))
	}
	if n.LinkDown {
		config = append(config, fmt.Sprintf("link_down=%d", btoi(n.LinkDown)))
	}
	if n.MacAddr != "" {
		config = append(config, fmt.Sprintf("macaddr=%s,%s=%s", n.MacAddr, string(n.Model), n.MacAddr))
	}
	if n.MTU != 0 {
		config = append(config, fmt.Sprintf("mtu=%d", n.MTU))
	}
	if n.Queues != 0 {
		config = append(config, fmt.Sprintf("queues=%d", n.Queues))
	}
	if n.Rate != "" {
		config = append(config, fmt.Sprintf("rate=%s", n.Rate))
	}
	if n.Tag != 0 {
		config = append(config, fmt.Sprintf("tag=%d", n.Tag))
	}
	if n.Trunks != nil {
		config = append(config, fmt.Sprintf("trunks=%s", strings.Join(itoaSlice(n.Trunks), ";")))
	}
	return strings.Join(config, ",")
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

// bool to int
func btoi(x bool) int8 {
	if x {
		return 1
	}
	return 0
}

// []int to []string
func itoaSlice(a []int) []string {
	b := []string{}
	for _, x := range a {
		b = append(b, fmt.Sprintf("%d", x))
	}
	return b
}
