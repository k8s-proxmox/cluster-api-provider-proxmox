package v1beta1

import (
	"fmt"

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
// type Image struct {
// 	// URL is a location of an image to deploy.
// 	URL string `json:"url"`

// 	// Checksum
// 	Checksum string `json:"checksum,omitempty"`

// 	// ChecksumType
// 	ChecksumType *string `json:"checksumType,omitempty"`
// }

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

// IPConfig
type IPConfig struct {
	IP       string `json:"ip,omitempty"`
	Gateway4 string `json:"gateway,omitempty"`
	DHCP     bool   `json:"dhcp,omitempty"`
}

// to do : user better logic
func (i *IPConfig) String() string {
	config := ""
	if i.IP != "" {
		config += fmt.Sprintf("ip=%s", i.IP)
	}
	if i.IP != "" && i.Gateway4 != "" {
		config += ","
	}
	if i.Gateway4 != "" {
		config += fmt.Sprintf("gw=%s", i.Gateway4)
	}
	if config == "" {
		config = "ip=dhcp"
	}
	return config
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
