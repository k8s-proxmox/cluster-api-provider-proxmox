/*
Copyright 2023 Teppei Sudo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"github.com/sp-yduck/proxmox-go/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// MachineFinalizer
	MachineFinalizer = "proxmoxmachine.infrastructure.cluster.x-k8s.io"
)

// ProxmoxMachineSpec defines the desired state of ProxmoxMachine
type ProxmoxMachineSpec struct {
	// ProviderID
	ProviderID *string `json:"providerID,omitempty"`

	// Node is proxmox node hosting vm instance which used for ProxmoxMachine
	Node string `json:"node,omitempty"`

	// Storage is name of proxmox storage used by this node.
	// The storage must support "images(VM Disks)" type of content.
	// cappx will use random storage if empty
	Storage string `json:"storage,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	// VMID is proxmox qemu's id
	VMID *int `json:"vmID,omitempty"`

	// Image is the image to be provisioned
	Image Image `json:"image"`

	// CloudInit defines options related to the bootstrapping systems where
	// CloudInit is used.
	CloudInit CloudInit `json:"cloudInit,omitempty"`

	// Hardware
	// +kubebuilder:default:={cpu:2,disk:"50G",memory:4096,networkDevice:{model:virtio,bridge:vmbr0,firewall:true}}
	Hardware Hardware `json:"hardware,omitempty"`

	// Network
	Network Network `json:"network,omitempty"`

	// Options for QEMU instance
	Options Options `json:"options,omitempty"`

	// FailureDomain is the failure domain unique identifier this Machine should be attached to, as defined in Cluster API.
	FailureDomain *string `json:"failureDomain,omitempty"`
}

// ProxmoxMachineStatus defines the observed state of ProxmoxMachine
type ProxmoxMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// FailureReason
	FailureReason *errors.MachineStatusError `json:"failureReason,omitempty"`

	// FailureMessage
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Addresses
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// Conditions
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`

	// Configuration
	Config api.VirtualMachineConfig `json:"config,omitempty"`

	// InstanceStatus is the status of the proxmox instance for this machine.
	// +optional
	InstanceStatus *InstanceStatus `json:"instanceStatus,omitempty"` // InstanceStatus
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this VSphereMachine belongs"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this ProxmoxMachine",priority=1
// +kubebuilder:printcolumn:name="VMID",type=string,JSONPath=`.spec.vmID`,priority=1
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.node`,priority=1
// +kubebuilder:printcolumn:name="Storage",type=string,JSONPath=`.spec.storage`,priority=1
// +kubebuilder:printcolumn:name="ProviderID",type=string,JSONPath=`.spec.providerID`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.instanceStatus`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Machine"

// ProxmoxMachine is the Schema for the proxmoxmachines API
type ProxmoxMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxmoxMachineSpec   `json:"spec,omitempty"`
	Status ProxmoxMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxmoxMachineList contains a list of ProxmoxMachine
type ProxmoxMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxmoxMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxmoxMachine{}, &ProxmoxMachineList{})
}
