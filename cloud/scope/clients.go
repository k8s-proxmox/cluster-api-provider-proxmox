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
	"github.com/sp-yduck/proxmox-go/proxmox"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

type ProxmoxServices struct {
	Compute *proxmox.Service
}

func newComputeService(ctx context.Context, serverRef infrav1.ServerRef, crClient client.Client) (*proxmox.Service, error) {
	secretRef := serverRef.SecretRef
	if secretRef == nil {
		return nil, errors.New("failed to get proxmox client form nil secretRef")
	}

	var secret corev1.Secret
	key := client.ObjectKey{Namespace: secretRef.Namespace, Name: secretRef.Name}
	if err := crClient.Get(ctx, key, &secret); err != nil {
		return nil, err
	}

	authConfig := proxmox.AuthConfig{
		Username: string(secret.Data["PROXMOX_USER"]),
		Password: string(secret.Data["PROXMOX_PASSWORD"]),
		TokenID:  string(secret.Data["PROXMOX_TOKENID"]),
		Secret:   string(secret.Data["PROXMOX_SECRET"]),
	}

	return proxmox.NewService(serverRef.Endpoint, authConfig, true)
}
