# qemu-scheduler

Scheduling refers to making sure that VM(QEMU) are matched to Proxmox Nodes.

## How qemu-scheduler select proxmox node to run qemu

Basic flow of the node selection process is `filter => score => select one node which has highest score`

### Filter Plugins

Filter plugins filter the node based on nodename, overcommit ratio etc.

#### regex plugin

Regex plugin is a one of the default Filter Plugin of qemu-scheduler. You can specify node name as regex format. 
```sh
key: node.qemu-scheduler/regex
value(example): node[0-9]+
```

### Score Plugins

Score plugins score the nodes based on resource etc.

## How to specify vmid
qemu-scheduler reads context and find key registerd to scheduler. If the context has any value of the registerd key, qemu-scheduler uses the plugin that matchies the key.

### Range Plugin
You can specify vmid range with `(start id)-(end id)` format.
```sh
key: vmid.qemu-scheduler/range
value(example): 100-150
```

### Regex Plugin
```sh
key: vmid.qemu-scheduler/regex
value(example): (12[0-9]|130)
```

## How qemu-scheduler works with CAPPX
CAPPX passes all the annotation (of `ProxmoxMachine`) key-values to scheduler's context. So if you will use Range Plugin for your `ProxmoxMachine`, your manifest must look like following.
```sh
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: ProxmoxMachine
metadata:
    name: sample-machine
    annotations:
        vmid.qemu-scheduler/range: 100-150 # this means your vmid will be chosen from the range of 100 to 150.
```

Also, you can specifies these annotations via `MachineDeployment` since Cluster API propagates some metadatas (ref: [metadata-propagation](https://cluster-api.sigs.k8s.io/developer/architecture/controllers/metadata-propagation.html#metadata-propagation)).

For example, your `MachineDeployment` may look like following.
```sh
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  annotations:
    caution: "# do not use here, because this annotation won't be propagated to your ProxmoxMachine"
  name: sample-machine-deployment
spec:
  template:
    metadata:
      annotations:
        node.qemu-scheduler/regex: node[0-9]+ # this annotation will be propagated to your ProxmoxMachine via MachineSet
```