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

package scope

import (
	"context"

	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/proxmox/pkg/service"
)

type ClusterScopeParams struct {
	ProxmoxServices
	Client         client.Client
	Cluster        *clusterv1.Cluster
	ProxmoxCluster *infrav1.ProxmoxCluster
}

func NewClusterScope(ctx context.Context, params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.ProxmoxCluster == nil {
		return nil, errors.New("failed to generate new scope from nil ProxmoxCluster")
	}

	if params.ProxmoxServices.Compute == nil {
		computeSvc, err := newComputeService(ctx, params.ProxmoxCluster.Spec.CredentialsRef, params.Client)
		if err != nil {
			return nil, errors.Errorf("failed to create proxmox compute client: %v", err)
		}
		params.ProxmoxServices.Compute = computeSvc
	}

	if params.ProxmoxServices.Remote == nil {
		remote, err := newRemoteClient(ctx, params.ProxmoxCluster.Spec.CredentialsRef, params.Client)
		if err != nil {
			return nil, errors.Errorf("failed to create remote client: %v", err)
		}
		params.ProxmoxServices.Remote = remote
	}

	helper, err := patch.NewHelper(params.ProxmoxCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterScope{
		client:          params.Client,
		Cluster:         params.Cluster,
		ProxmoxCluster:  params.ProxmoxCluster,
		ProxmoxServices: params.ProxmoxServices,
		patchHelper:     helper,
	}, err
}

type ClusterScope struct {
	ProxmoxServices
	client         client.Client
	patchHelper    *patch.Helper
	Cluster        *clusterv1.Cluster
	ProxmoxCluster *infrav1.ProxmoxCluster
}

func (s *ClusterScope) Name() string {
	return s.Cluster.Name
}

func (s *ClusterScope) Namespace() string {
	return s.Cluster.Namespace
}

func (s *ClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	return s.ProxmoxCluster.Spec.ControlPlaneEndpoint
}

func (s *ClusterScope) Storage() infrav1.Storage {
	return s.ProxmoxCluster.Spec.Storage
}

func (s *ClusterScope) CloudClient() *service.Service {
	return s.ProxmoxServices.Compute
}

func (s *ClusterScope) RemoteClient() *SSHClient {
	return s.ProxmoxServices.Remote
}

func (s *ClusterScope) Close() error {
	return s.PatchObject()
}

func (s *ClusterScope) SetReady() {
	s.ProxmoxCluster.Status.Ready = true
}

func (s *ClusterScope) SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint) {
	s.ProxmoxCluster.Spec.ControlPlaneEndpoint = endpoint
}

func (s *ClusterScope) SetStorage(storage infrav1.Storage) {
	s.ProxmoxCluster.Spec.Storage = storage
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterScope) PatchObject() error {
	return s.patchHelper.Patch(context.TODO(), s.ProxmoxCluster)
}
