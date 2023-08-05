package v1beta1

import (
	"fmt"
	"strings"

	"github.com/sp-yduck/proxmox-go/api"
)

// ServerRef is used for configuring Proxmox client
type ServerRef struct {
	// endpoint is the address of the Proxmox-VE REST API endpoint.
	Endpoint string `json:"endpoint"`

	// to do : login type should be an option
	// user&pass or token

	// to do : client options like insecure tls verify

	// SecretRef is a reference for secret which contains proxmox login secrets
	// and ssh configs for proxmox nodes
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
	// URL is a location of an image to deploy.
	URL string `json:"url"`

	// Checksum
	Checksum string `json:"checksum,omitempty"`

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

	// hard disk size
	// +kubebuilder:default:="50G"
	Disk string `json:"disk,omitempty"`
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

// IPConfig defines IP addresses and gateways for corresponding interface
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

type InstanceStatus string

var (
	InstanceStatusPaused  = InstanceStatus(api.ProcessStatusPaused)
	InstanceStatusRunning = InstanceStatus(api.ProcessStatusRunning)
	InstanceStatusStopped = InstanceStatus(api.ProcessStatusStopped)
)
