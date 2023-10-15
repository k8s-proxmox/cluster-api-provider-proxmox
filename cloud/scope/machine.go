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
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/providerid"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler"
)

type MachineScopeParams struct {
	ProxmoxServices
	Client         client.Client
	Machine        *clusterv1.Machine
	ProxmoxMachine *infrav1.ProxmoxMachine
	ClusterGetter  *ClusterScope
	Scheduler      *scheduler.Scheduler
}

func NewMachineScope(params MachineScopeParams) (*MachineScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachineScope")
	}
	if params.Machine == nil {
		return nil, errors.New("failed to generate new scope from nil Machine")
	}
	if params.ProxmoxMachine == nil {
		return nil, errors.New("failed to generate new scope from nil ProxmoxMachine")
	}
	if params.ClusterGetter == nil {
		return nil, errors.New("failed to generate new scope form nil ClusterScope")
	}
	if params.Scheduler == nil {
		return nil, errors.New("failed to generate new scope form nil Scheduler")
	}

	helper, err := patch.NewHelper(params.ProxmoxMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &MachineScope{
		client:         params.Client,
		Machine:        params.Machine,
		ProxmoxMachine: params.ProxmoxMachine,
		patchHelper:    helper,
		ClusterGetter:  params.ClusterGetter,
		Scheduler:      params.Scheduler,
	}, err
}

type MachineScope struct {
	client         client.Client
	patchHelper    *patch.Helper
	Machine        *clusterv1.Machine
	ProxmoxMachine *infrav1.ProxmoxMachine
	ClusterGetter  *ClusterScope
	Scheduler      *scheduler.Scheduler
}

func (m *MachineScope) CloudClient() *proxmox.Service {
	return m.ClusterGetter.CloudClient()
}

func (m *MachineScope) GetScheduler() *scheduler.Scheduler {
	return m.Scheduler
}

func (m *MachineScope) GetStorage() infrav1.Storage {
	return m.ClusterGetter.ProxmoxCluster.Spec.Storage
}

func (m *MachineScope) Name() string {
	return m.ProxmoxMachine.Name
}

func (m *MachineScope) Namespace() string {
	return m.ProxmoxMachine.Namespace
}

func (m *MachineScope) NodeName() string {
	return m.ProxmoxMachine.Spec.Node
}

func (m *MachineScope) SetNodeName(name string) {
	m.ProxmoxMachine.Spec.Node = name
}

// func (m *MachineScope) Client() Compute {
// 	return m.ClusterGetter.Client()
// }

func (m *MachineScope) GetBootstrapData() (string, error) {
	if m.Machine.Spec.Bootstrap.DataSecretName == nil {
		return "", errors.New("error retrieving bootstrap data: linked Machine's bootstrap.dataSecretName is nil")
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: m.Namespace(), Name: *m.Machine.Spec.Bootstrap.DataSecretName}
	if err := m.client.Get(context.TODO(), key, secret); err != nil {
		return "", errors.Wrapf(err, "failed to retrieve bootstrap data secret for ProxmoxMachine %s/%s", m.Namespace(), m.Name())
	}

	value, ok := secret.Data["value"]
	if !ok {
		return "", errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return string(value), nil
}

func (m *MachineScope) Close() error {
	return m.PatchObject()
}

func (m *MachineScope) GetInstanceStatus() *infrav1.InstanceStatus {
	return m.ProxmoxMachine.Status.InstanceStatus
}

// SetInstanceStatus sets the ProxmoxMachine instance status.
func (m *MachineScope) SetInstanceStatus(v infrav1.InstanceStatus) {
	m.ProxmoxMachine.Status.InstanceStatus = &v
}

func (m *MachineScope) GetBiosUUID() *string {
	parsed, err := noderefutil.NewProviderID(m.GetProviderID()) //nolint: staticcheck
	if err != nil {
		return nil
	}
	return pointer.String(parsed.ID()) //nolint: staticcheck
}

func (m *MachineScope) GetProviderID() string {
	if m.ProxmoxMachine.Spec.ProviderID != nil {
		return *m.ProxmoxMachine.Spec.ProviderID
	}
	return ""
}

func (m *MachineScope) GetVMID() *int {
	return m.ProxmoxMachine.Spec.VMID
}

func (m *MachineScope) GetImage() infrav1.Image {
	return m.ProxmoxMachine.Spec.Image
}

func (m *MachineScope) GetCloudInit() infrav1.CloudInit {
	return m.ProxmoxMachine.Spec.CloudInit
}

func (m *MachineScope) GetNetwork() infrav1.Network {
	return m.ProxmoxMachine.Spec.Network
}

func (m *MachineScope) GetHardware() infrav1.Hardware {
	// set default value if empty
	if m.ProxmoxMachine.Spec.Hardware.CPU == 0 {
		m.ProxmoxMachine.Spec.Hardware.CPU = 2
	}
	if m.ProxmoxMachine.Spec.Hardware.Memory == 0 {
		m.ProxmoxMachine.Spec.Hardware.Memory = 4096
	}
	if m.ProxmoxMachine.Spec.Hardware.Disk == "" {
		m.ProxmoxMachine.Spec.Hardware.Disk = "50G"
	}
	return m.ProxmoxMachine.Spec.Hardware
}

func (m *MachineScope) GetOptions() infrav1.Options {
	return m.ProxmoxMachine.Spec.Options
}

// SetProviderID sets the ProxmoxMachine providerID in spec.
func (m *MachineScope) SetProviderID(uuid string) error {
	providerid, err := providerid.New(uuid)
	if err != nil {
		return err
	}
	m.ProxmoxMachine.Spec.ProviderID = pointer.String(providerid.String())
	return nil
}

func (m *MachineScope) SetVMID(vmid int) {
	m.ProxmoxMachine.Spec.VMID = &vmid
}

func (m *MachineScope) SetConfigStatus(config api.VirtualMachineConfig) {
	m.ProxmoxMachine.Status.Config = config
}

func (m *MachineScope) SetReady() {
	m.ProxmoxMachine.Status.Ready = true
}

func (m *MachineScope) SetFailureMessage(v error) {
	m.ProxmoxMachine.Status.FailureMessage = pointer.String(v.Error())
}

func (m *MachineScope) SetFailureReason(v capierrors.MachineStatusError) {
	m.ProxmoxMachine.Status.FailureReason = &v
}

// PatchObject persists the cluster configuration and status.
func (s *MachineScope) PatchObject() error {
	return s.patchHelper.Patch(context.TODO(), s.ProxmoxMachine)
}
