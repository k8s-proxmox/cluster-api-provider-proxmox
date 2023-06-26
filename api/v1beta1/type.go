package v1beta1

import (
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
