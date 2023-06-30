# cluster-api-provider-proxmox (CAPP)

cluster-api-provider-proxmox is a Cluster API [infrastructure provider](https://cluster-api.sigs.k8s.io/developer/providers/cluster-infrastructure.html) implementation for [Proxmox VE](https://pve.proxmox.com/wiki/Main_Page).

## Description
cluster-api-provider-proxmox provides only infrastructure controller (`ProxmoxCluster` and `ProxmoxMachine`). To bootstrap your cluster/machine you need to provide [Control Plane provider](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/control-plane.html#crd-contracts) and [Bootstrap provider](https://cluster-api.sigs.k8s.io/developer/providers/bootstrap.html). For example [KubeadmControlPlane](https://github.com/kubernetes-sigs/cluster-api/tree/main/controlplane/kubeadm) and [KubeadmBootstrap](https://github.com/kubernetes-sigs/cluster-api/tree/main/bootstrap/kubeadm).

## Quick Start
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Initialize Management cluster

for more information : https://cluster-api.sigs.k8s.io/user/quick-start.html#initialize-the-management-cluster

```sh
# install cluster-api crd including controlplane provider & bootstrap provider
clusterctl init

# install cluster-api-provider-proxmox crd & controller
make deploy
```
2. Create your first workload cluster
```sh
# cluster & infra cluster
kubectl apply -f config/samples/cluster.yaml
kubectl apply -f config/samples/infrastructure_v1beta1_proxmoxcluster.yaml

# controlplane
kubectl apply -f config/samples/controlplane.yaml

# machine & bootstrap & infra machine
kubectl apply -f config/samples/machine.yaml
kubectl apply -f config/samples/bootstrap.yaml
kubectl apply -f config/samples/infrastructure_v1beta1_proxmoxcluster.yaml

# proxmox configs
kubetl apply -f <your-proxmox-config-secret>.yaml
```

You need to provide your proxmox information through secret. 
```yaml
# <your-proxmox-config-secret>.yaml
apiVersion: v1
data:
  # for proxmox API
  PROXMOX_PASSWORD: "<base 64>"
  PROXMOX_USER: "<base 64>"
  # for ssh into the node to bootstrapping VMs
  ## * current CAPP is compatible with only single node proxmox cluster
  NODE_URL: "<base 64>"
  NODE_USER: "<base 64>"
  NODE_PASSWORD: "<base 64>"
kind: Secret
metadata:
  name: proxmoxcluster-sample
type: Opaque
```

## Fetures

- No need to prepare vm templates. You can specify any vm image in `ProxmoxMachine.Spec.Image`.

- Supports custom cloud-config (user data). CAPP uses ssh for bootstrapping nodes so it can applies custom cloud-config that can not be achieved by only Proxmox API.

## Compatibility

CAPP is tested with `pve-manager/7.4-3/9002ab8a (running kernel: 5.15.102-1-pve)`

## How it works
This project aims to follow the Cluster API [Provider contract](https://cluster-api.sigs.k8s.io/developer/providers/contracts.html).

## Contributing

Are you interested in contributing to cluster-api-provider-proxmox? Do not hesitate to open Github issues. 

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

