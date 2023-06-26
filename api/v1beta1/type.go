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

// IPConfig
type IPConfig struct {
	IP       string `json:"ip,omitempty"`
	Gateway4 string `json:"gateway,omitempty"`
}

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
	return config
}

// Storage for image and snippets
type Storage struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

type InstanceStatus string

var (
	InstanceStatusPaused  = InstanceStatus(vm.ProcessStatusPaused)
	InstanceStatusRunning = InstanceStatus(vm.ProcessStatusRunning)
	InstanceStatusStopped = InstanceStatus(vm.ProcessStatusStopped)
)
