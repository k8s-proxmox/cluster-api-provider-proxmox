# cluster-api-provider-proxmox (CAPP)

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
# install cluster-api crd including controlplane provider & bootstrap provider
clusterctl init

# install cluster-api-provider-proxmox crd & controller
make deploy
```
**Note:** container images are available at [here](https://hub.docker.com/r/spyduck/cluster-api-provider-proxmox/tags)

2. Create your first workload cluster
```sh
# export env variables
export CONTROLPLANE_HOST=X.X.X.X
export PROXMOX_URL=X.X.X.X:8006
export GATEWAY_ADDRESS=X.X.X.X
export NAMESERVER_ADDRESS=X.X.X.X
export PROXMOX_PASSWORD_BASE64=$(echo -n <password> | base64)
export PROXMOX_USER_BASE64=$(echo -n <user@pam> | base64)
export NODE_URL_BASE64=$(echo -n <node.ssh.url:22> | base64)
export NODE_USER_BASE64=$(echo -n <node-ssh-user> | base64)
export NODE_PASSWORD_BASE64=$(echo -n <node-ssh-password> | base64)

make create-workload-cluster
```

3. Access your first workload cluster !!

Usually it takes 2~5 mins to complete bootstrap the nodes.
```sh
# get workload cluster's kubeconfig
clusterctl get kubeconfig cappx-test > kubeconfig.yaml

# get node command for workload cluster
kubectl --kubeconfig=kubeconfig.yaml get node
```

4. Tear down your workload cluster
```sh
make delete-workload-cluster
```

## Fetures

- No need to prepare vm templates. You can specify any vm image in `ProxmoxMachine.Spec.Image`.

- Supports custom cloud-config (user data). CAPP uses ssh for bootstrapping nodes so it can applies custom cloud-config that can not be achieved by only Proxmox API.

## Compatibility

### Proxmox-VE REST API

CAPP is tested with `pve-manager/7.4-3/9002ab8a (running kernel: 5.15.102-1-pve)`.

### Cluster API

|                       | Cluster API v1alpha4 | Cluster API v1beta1 |
| --------------------- | :------------------: | :-----------------: |
| CAPP v1beta1 `(v0.x)` |          ?           |          ✓          |

### ControlPlane & Bootstrap provider 

CAPP is tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## How it works
This project aims to follow the Cluster API [Provider contract](https://cluster-api.sigs.k8s.io/developer/providers/contracts.html).

### ProxmoxCluster

Because Proxmox-VE does not provide LBaaS solution, CAPP does not follow the [typical infra-cluster logic](https://cluster-api.sigs.k8s.io/developer/providers/cluster-infrastructure.html#behavior). ProxmoxCluster controller reconciles only Proxmox storages used for instances. You need to prepare control plane load balancer by yourself if you creates HA control plane workload cluster.

### ProxmoxMachine

ProxmoxMachine controller follows the [typical infra-machine logic](https://cluster-api.sigs.k8s.io/developer/providers/machine-infrastructure.html#behavior). To bootstrap your machine, CAPP supports only `cloud-config` type bootstrap data secret. CAPP is mainly tested with [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

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

