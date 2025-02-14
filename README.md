# cluster-api-provider-proxmox (CAPPX)

⚠️ **Note:** Ongoing development has moved to the [IONOS Proxmox Provider](https://github.com/ionos-cloud/cluster-api-provider-proxmox).

## [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/k8s-proxmox/cluster-api-provider-proxmox?sort=semver)](https://github.com/k8s-proxmox/cluster-api-provider-proxmox/releases/latest) [![Go Report Card](https://goreportcard.com/badge/github.com/k8s-proxmox/cluster-api-provider-proxmox)](https://goreportcard.com/report/github.com/k8s-proxmox/cluster-api-provider-proxmox) [![CI](https://github.com/k8s-proxmox/cluster-api-provider-proxmox/actions/workflows/ci.yaml/badge.svg)](https://github.com/k8s-proxmox/cluster-api-provider-proxmox/actions/workflows/ci.yaml) [![GitHub license](https://img.shields.io/github/license/k8s-proxmox/cluster-api-provider-proxmox)](https://github.com/k8s-proxmox/cluster-api-provider-proxmox/blob/main/LICENSE)

cluster-api-provider-proxmox is a Cluster API [infrastructure provider](https://cluster-api.sigs.k8s.io/developer/providers/cluster-infrastructure.html) implementation for [Proxmox VE](https://pve.proxmox.com/wiki/Main_Page).

## Description

cluster-api-provider-proxmox provides only infrastructure controller (`ProxmoxCluster` and `ProxmoxMachine`). To bootstrap your cluster/machine you need to provide [Control Plane provider](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/control-plane.html#crd-contracts) and [Bootstrap provider](https://cluster-api.sigs.k8s.io/developer/providers/bootstrap.html). For example [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## Quick Start

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### with KubeadmControlplane/Bootstrap

1. Initialize Management cluster

for more information : https://cluster-api.sigs.k8s.io/user/quick-start.html#initialize-the-management-cluster

```sh
# install cluster-api components
export EXP_CLUSTER_RESOURCE_SET=true
clusterctl init --infrastructure=proxmox:v0.4.3 --config https://raw.githubusercontent.com/k8s-proxmox/cluster-api-provider-proxmox/main/clusterctl.yaml
```

**Note:** container images are available at [ghcr.io/k8s-proxmox/cluster-api-provider-proxmox:\<tag\>](https://github.com/k8s-proxmox/cluster-api-provider-proxmox/pkgs/container/cluster-api-provider-proxmox)

2. Create your first workload cluster

```sh
# export env variables
export CONTROLPLANE_HOST=X.X.X.X                   # control-plane vip
export PROXMOX_URL=https://X.X.X.X:8006/api2/json
export PROXMOX_PASSWORD=password
export PROXMOX_USER=user@pam

# generate manifests (available flags: --target-namespace, --kubernetes-version, --control-plane-machine-count, --worker-machine-count)
clusterctl generate cluster cappx-test --control-plane-machine-count=3 --infrastructure=proxmox:v0.4.3 --config https://raw.githubusercontent.com/k8s-proxmox/cluster-api-provider-proxmox/main/clusterctl.yaml > cappx-test.yaml

# inspect and edit
vi cappx-test.yaml

# apply manifests
kubectl apply -f cappx-test.yaml
```

3. Access your first workload cluster !!

Usually it takes 2~10 mins to complete bootstrapping the nodes.

```sh
# get workload cluster's kubeconfig
clusterctl get kubeconfig cappx-test > kubeconfig.yaml

# get node command for workload cluster
kubectl --kubeconfig=kubeconfig.yaml get node
### example output: this is your first workload cluster !!
## NAME                            STATUS     ROLES           AGE     VERSION
## cappx-test-controlplane-qc9vw   NotReady   control-plane   6m53s   v1.27.3
```

3-a. [OPTIONAL] Apply your favorite CNI to your workload cluster

Until you apply CNI to your cluster, all the node is NotReady. After this step, all your nodes will become Ready :)

```sh
# use weave-cni for this example
kubectl --kubeconfig=kubeconfig.yaml apply -f https://github.com/weaveworks/weave/releases/download/v2.8.1/weave-daemonset-k8s.yaml
```

4. Tear down your workload cluster

```sh
kubectl delete cluster cappx-test
```

## Fetures

- No need to prepare vm templates. You can specify any vm image in `ProxmoxMachine.Spec.Image`. CAPPX bootstrap your vm from scratch. (Supports `iso` type of image format.)

- Supports custom cloud-config (user data). CAPPX uses VNC websockert for bootstrapping nodes so it can applies custom cloud-config that can not be achieved by only Proxmox API.

- Flexible vmid/node assigning. You can flexibly assign vmid to your qemu and flexibly schedule qemus to proxmox nodes. For more details please check [qemu-scheduler](./cloud/scheduler/).

### Node Images

CAPPX is compatible with `iso` format of image. You can build your own node image and use it for `ProxmoxMachine`.

CAPPX relies on a few prerequisites which have to be already installed in the used operating system images, e.g. a container runtime, kubelet, kubeadm,.. .

To build your custom node image, you can use [kubernetes-sigs/image-builder](https://github.com/kubernetes-sigs/image-builder) project.

If it isn't possible to pre-install those prerequisites in the image, you can always deploy and execute some custom scripts through the `ProxmoxMachine.spec.cloudInit` or `KubeadmConfig`. Example MD can be found [ubuntu2204.yaml](examples/machine_deployment/ubuntu2204.yaml).

## Compatibility

### Proxmox-VE REST API

CAPPX is tested with `pve-manager/7.4-3/9002ab8a (running kernel: 5.15.102-1-pve)`.

### Cluster API

|                        | Cluster API v1alpha4 | Cluster API v1beta1 |
| ---------------------- | :------------------: | :-----------------: |
| CAPPX v1beta1 `(v0.x)` |          ?           |          ✓          |

### ControlPlane & Bootstrap provider

CAPPX is tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## How it works

This project aims to follow the Cluster API [Provider contract](https://cluster-api.sigs.k8s.io/developer/providers/contracts.html).

### ProxmoxCluster

Because Proxmox-VE does not provide LBaaS solution, CAPPX does not follow the [typical infra-cluster logic](https://cluster-api.sigs.k8s.io/developer/providers/cluster-infrastructure.html#behavior). ProxmoxCluster controller reconciles only Proxmox storages used for instances. You need to prepare control plane load balancer by yourself if you creates HA control plane workload cluster. In the [cluster-template.yaml](./templates/cluster-template.yaml), you can find HA control plane example with [kube-vip](https://github.com/kube-vip/kube-vip).

### ProxmoxMachine

ProxmoxMachine controller follows the [typical infra-machine logic](https://cluster-api.sigs.k8s.io/developer/providers/machine-infrastructure.html#behavior). To bootstrap your machine, CAPPX supports only `cloud-config` type bootstrap data secret. CAPPX is mainly tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## Development

### Testing

#### Unit Testing

```sh
make unit-test
```

#### Unit and Integration Testing

```sh
export PROXMOX_URL=https://X.X.X.X:8006/api2/json
export PROXMOX_PASSWORD=password
export PROXMOX_USER=user@pam

make test
```

#### E2E Testing

```sh
export CONTROLPLANE_HOST=X.X.X.X
export PROXMOX_URL=https://X.X.X.X:8006/api2/json
export PROXMOX_PASSWORD=password
export PROXMOX_USER=user@pam

make e2e
```

## Contributing

Are you interested in contributing to cluster-api-provider-proxmox? Do not hesitate to open GitHub issues.

## License

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
