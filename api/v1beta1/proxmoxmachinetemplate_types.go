/*
Copyright 2023 Simplysoft GmbH.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// ProxmoxMachineTemplateSpec defines the desired state of ProxmoxMachineTemplate
type ProxmoxMachineTemplateSpec struct {
	Template ProxmoxMachineTemplateSpecTemplate `json:"template"`

	// VM ID Range that will be used for individual machines
	// +optional
	VMIDs *ProxmoxMachineTemplateVmIdRange `json:"vmIDs,omitempty"`

	// Restrict template to specific proxmox nodes. When failure domains are enabled, they will have
	// priority the configured nodes in the template
	// +optional
	Nodes []string `json:"nodes,omitempty"`
}

type ProxmoxMachineTemplateSpecTemplate struct {
	// +optional
	ObjectMeta clusterv1.ObjectMeta                   `json:"metadata.omitempty"`
	Spec       ProxmoxMachineTemplateSpecTemplateSpec `json:"spec"`
}

type ProxmoxMachineTemplateSpecTemplateSpec struct {
	// Image is the image to be provisioned
	Image Image `json:"image"`

	// CloudInit defines options related to the bootstrapping systems where
	// CloudInit is used.
	// +optional
	CloudInit CloudInit `json:"cloudInit,omitempty"`

	// Hardware
	Hardware Hardware `json:"hardware,omitempty"`

	// Network
	Network Network `json:"network,omitempty"`

	// Options
	// +optional
	Options Options `json:"options,omitempty"`
}

type ProxmoxMachineTemplateVmIdRange struct {
	// Start of VM ID range
	Start int `json:"start"`

	// End of VM ID range
	// +optional
	End int `json:"end,omitempty"`
}

// ProxmoxMachineTemplateStatus defines the observed state of ProxmoxMachineTemplate
type ProxmoxMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProxmoxMachineTemplate is the Schema for the proxmoxmachinetemplates API
type ProxmoxMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxmoxMachineTemplateSpec   `json:"spec,omitempty"`
	Status ProxmoxMachineTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxmoxMachineTemplateList contains a list of ProxmoxMachineTemplate
type ProxmoxMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxmoxMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxmoxMachineTemplate{}, &ProxmoxMachineTemplateList{})
}
