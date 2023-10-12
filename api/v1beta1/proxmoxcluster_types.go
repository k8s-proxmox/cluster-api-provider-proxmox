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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer
	ClusterFinalizer       = "proxmoxcluster.infrastructure.cluster.x-k8s.io"
	ClusterSecretFinalizer = ClusterFinalizer + "/secret"
)

// ProxmoxClusterSpec defines the desired state of ProxmoxCluster
type ProxmoxClusterSpec struct {
	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// ServerRef is used for configuring Proxmox client
	ServerRef ServerRef `json:"serverRef"`

	// storage is for proxmox storage used by vm instances
	// +optional
	Storage Storage `json:"storage"`

	// Nodes to be used for vm instances
	// +optional
	Nodes []string `json:"nodes,omitempty"`

	// Provide config to enable failure domains
	// +optional
	FailureDomainConfig *ClusterFailureDomainConfig `json:"failureDomain,omitempty"`

	// Proxmox resource pool to be used for all vms created
	// +optional
	ResourcePool string `json:"resourcePool,omitempty"`
}

// ProxmoxClusterStatus defines the observed state of ProxmoxCluster
type ProxmoxClusterStatus struct {
	// Ready
	Ready bool `json:"ready"`

	// FailureDomains
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`

	// Conditions
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for ProxmoxMachine"
// +kubebuilder:printcolumn:name="Proxmox-Server",type="string",JSONPath=".spec.serverRef.endpoint",description="Server is the address of the Proxmox API endpoint."
// +kubebuilder:printcolumn:name="ControlPlane",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="kube-apiserver Endpoint"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Machine"

// ProxmoxCluster is the Schema for the proxmoxclusters API
type ProxmoxCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxmoxClusterSpec   `json:"spec,omitempty"`
	Status ProxmoxClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxmoxClusterList contains a list of ProxmoxCluster
type ProxmoxClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxmoxCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxmoxCluster{}, &ProxmoxClusterList{})
}
