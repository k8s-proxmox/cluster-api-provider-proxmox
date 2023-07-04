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
)

type ProxmoxServices struct {
	Compute *service.Service
	Remote  *SSHClient
}

func newComputeService(ctx context.Context, serverRef infrav1.ServerRef, crClient client.Client) (*service.Service, error) {
	secretRef := serverRef.SecretRef
	if secretRef == nil {
		return nil, errors.New("failed to get proxmox client form nil secretRef")
	}

	var secret corev1.Secret
	key := client.ObjectKey{Namespace: secretRef.Namespace, Name: secretRef.Name}
	if err := crClient.Get(ctx, key, &secret); err != nil {
		return nil, err
	}

	proxmoxUser, ok := secret.Data["PROXMOX_USER"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_USER from Secret : %v", key)
	}
	proxmoxPassword, ok := secret.Data["PROXMOX_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_PASSWORD from Secret : %v", key)
	}

	return service.NewServiceWithLogin(serverRef.Endpoint, string(proxmoxUser), string(proxmoxPassword))
}

func newRemoteClient(ctx context.Context, secretRef *infrav1.ObjectReference, crClient client.Client) (*SSHClient, error) {
	if secretRef == nil {
		return nil, errors.New("failed to get proxmox client form nil secretRef")
	}

	var secret corev1.Secret
	key := client.ObjectKey{Namespace: secretRef.Namespace, Name: secretRef.Name}
	if err := crClient.Get(ctx, key, &secret); err != nil {
		return nil, err
	}

	nodeurl, ok := secret.Data["NODE_URL"]
	if !ok {
		return nil, errors.Errorf("failed to fetch NODE_URL from Secret : %v", key)
	}
	nodeuser, ok := secret.Data["NODE_USER"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_USER from Secret : %v", key)
	}
	nodepassword, ok := secret.Data["NODE_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("failed to fetch PROXMOX_PASSWORD from Secret : %v", key)
	}

	return NewSSHClient(string(nodeurl), string(nodeuser), string(nodepassword))
}
