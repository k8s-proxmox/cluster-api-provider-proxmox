# cluster-api-provider-proxmox (CAPPX)

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
clusterctl init --infrastructure=proxmox:v0.2.2 --config https://raw.githubusercontent.com/sp-yduck/cluster-api-provider-proxmox/main/clusterctl.yaml
```
**Note:** container images are available at [here](https://hub.docker.com/r/spyduck/cluster-api-provider-proxmox/tags)

2. Create your first workload cluster
```sh
# export env variables
export CONTROLPLANE_HOST=X.X.X.X   # for control-plane node
export GATEWAY_ADDRESS=X.X.X.X     # for control-plane node
export NAMESERVER_ADDRESS=X.X.X.X  # for contro-lplane node
export PROXMOX_URL=X.X.X.X:8006
export PROXMOX_PASSWORD=password
export PROXMOX_USER=user@pam
export NODE_URL=node.ssh.url:22
export NODE_USER=node-ssh-user
export NODE_PASSWORD=node-ssh-password

# generate manifests
clusterctl generate cluster cappx-test --infrastructure=proxmox:v0.2.3 --config https://raw.githubusercontent.com/sp-yduck/cluster-api-provider-proxmox/main/clusterctl.yaml > cappx-test.yaml

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
```

4. Tear down your workload cluster
```sh
kubectl delete cluster cappx-test
```

## Fetures

- No need to prepare vm templates. You can specify any vm image in `ProxmoxMachine.Spec.Image`.

- Supports custom cloud-config (user data). CAPPX uses ssh for bootstrapping nodes so it can applies custom cloud-config that can not be achieved by only Proxmox API.

### Node Images

CAPPX is compatible with `qcow2` image. You can build your own node image and use it for `ProxmoxMachine`.

CAPPX relies on a few prerequisites which have to be already installed in the used operating system images, e.g. a container runtime, kubelet, kubeadm,.. .

To build your custom node image, you can use [kubernetes-sigs/image-builder](https://github.com/kubernetes-sigs/image-builder) project. 

Also there are some available out-of-box images published other communities such as [Metal3](https://github.com/metal3-io). For example https://artifactory.nordix.org/ui/native/metal3/images/. Example MD can be found [metal3-ubuntu2204-k8s127.yaml](examples/machine_deployment/metal3-ubuntu2204-k8s127.yaml).

If it isn't possible to pre-install those prerequisites in the image, you can always deploy and execute some custom scripts through the `ProxmoxMachine.spec.cloudInit` or `KubeadmConfig.spec.preKubeadmCommands` etc. . Example MD can be found [ubuntu2204.yaml](examples/machine_deployment/ubuntu2204.yaml).

## Compatibility

### Proxmox-VE REST API

CAPPX is tested with `pve-manager/7.4-3/9002ab8a (running kernel: 5.15.102-1-pve)`.

### Cluster API

|                       | Cluster API v1alpha4 | Cluster API v1beta1 |
| --------------------- | :------------------: | :-----------------: |
| CAPPX v1beta1 `(v0.x)` |          ?           |          ✓          |

### ControlPlane & Bootstrap provider 

CAPPX is tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## How it works
This project aims to follow the Cluster API [Provider contract](https://cluster-api.sigs.k8s.io/developer/providers/contracts.html).

### ProxmoxCluster

Because Proxmox-VE does not provide LBaaS solution, CAPPX does not follow the [typical infra-cluster logic](https://cluster-api.sigs.k8s.io/developer/providers/cluster-infrastructure.html#behavior). ProxmoxCluster controller reconciles only Proxmox storages used for instances. You need to prepare control plane load balancer by yourself if you creates HA control plane workload cluster.

### ProxmoxMachine

ProxmoxMachine controller follows the [typical infra-machine logic](https://cluster-api.sigs.k8s.io/developer/providers/machine-infrastructure.html#behavior). To bootstrap your machine, CAPPX supports only `cloud-config` type bootstrap data secret. CAPPX is mainly tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

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

