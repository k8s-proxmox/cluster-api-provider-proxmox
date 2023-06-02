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
	"github.com/sp-yduck/proxmox/pkg/service"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	// "github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services/compute/vm"
)

type ProxmoxServices struct {
	Compute *service.Service
}

func newComputeService(ctx context.Context, credentialsRef *infrav1.ObjectReference, crClient client.Client) (*service.Service, error) {
	var secret corev1.Secret
	key := client.ObjectKey{Namespace: credentialsRef.Namespace, Name: credentialsRef.Name}
	if err := crClient.Get(ctx, key, &secret); err != nil {
		return nil, err
	}

	proxmoxUrl, ok := secret.Data["PROXMOX_URL"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_URL from Secret : %v", key)
	}
	proxmoxUser, ok := secret.Data["PROXMOX_USER"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_USER from Secret : %v", key)
	}
	proxmoxPassword, ok := secret.Data["PROXMOX_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_PASSWORD from Secret : %v", key)
	}

	return service.NewServiceWithLogin(string(proxmoxUrl), string(proxmoxUser), string(proxmoxPassword))
}
