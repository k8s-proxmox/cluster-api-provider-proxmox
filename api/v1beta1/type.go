package v1beta1

import (
	"fmt"
	"strings"

	"github.com/sp-yduck/proxmox/pkg/service/node/vm"
)

// ObjectReference is a reference to another Kubernetes object instance.
type ObjectReference struct {
	// Namespace of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
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
	// number of CPU cores : 1 ~
	// +kubebuilder:validation:Mimimum:=1
	// +kubebuilder:default:=2
	CPU int `json:"cpu,omitempty"`

	// amount of RAM for the VM in MiB : 16 ~
	// +kubebuilder:validation:Minimum:=16
	// +kubebuilder:default:=4096
	Memory int `json:"memoty,omitempty"`
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
	InstanceStatusPaused  = InstanceStatus(vm.ProcessStatusPaused)
	InstanceStatusRunning = InstanceStatus(vm.ProcessStatusRunning)
	InstanceStatusStopped = InstanceStatus(vm.ProcessStatusStopped)
)
