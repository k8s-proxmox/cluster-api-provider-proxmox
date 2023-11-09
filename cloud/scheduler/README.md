# qemu-scheduler

Scheduling refers to making sure that VM(QEMU) are matched to Proxmox Nodes.

## How qemu-scheduler select proxmox node to run qemu

Basic flow of the node selection process is `filter => score => select one node which has highest score`

### Filter Plugins

Filter plugins filter the node based on nodename, overcommit ratio etc. So that we can avoid to run qemus on not desired Proxmox nodes.

- [NodeName plugin](./plugins/nodename/node_name.go) (pass the node matching specified node name)
- [CPUOvercommit plugin](./plugins/overcommit/cpu_overcommit.go) (pass the node that has enough cpu against running vm)
- [MemoryOvercommit plugin](./plugins/overcommit/memory_overcommit.go) (pass the node that has enough memory against running vm)
- [NodeRegex plugin](./plugins/regex/node_regex.go) (pass the node matching specified regex)

#### regex plugin

Regex plugin is a one of the default Filter Plugin of qemu-scheduler. You can specify node name as regex format. 
```sh
key: node.qemu-scheduler/regex
value(example): node[0-9]+
```

### Score Plugins

Score plugins score the nodes based on resource etc. So that we can run qemus on the most appropriate Proxmox node.

- [NodeResource plugin](./plugins/noderesource/node_resrouce.go) (nodes with more resources have higher scores)
- [Random plugin](./plugins/random/random.go) (diabled by default. just a reference implementation of score plugin)

## How to specify vmid
qemu-scheduler reads context and find key registerd to scheduler. If the context has any value of the registerd key, qemu-scheduler uses the plugin that matchies the key.

- [Range plugin](./plugins/idrange/idrange.go) (select minimum availabe vmid from the specified id range)
- [VMIDRegex plugin](./plugins/regex/vmid_regex.go) (select minimum availabe vmid matching specified regex)

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

## How to configure (or disable/enable) specific Plugins

By default, all the plugins are enabled. You can disable specific plugins via plugin-config. for CAPPX, check example ConfigMap [here](../../config/manager/manager.yaml)
```sh
# example plugin-config.yaml

# plugin type name (scores, filters, vmids)
filters:
  CPUOvercommit:
    enable: false # disable
  MemoryOvercommit:
    enable: true   # enable (can be omitted)
vmids:
  Regex:
    enable: false # disable
```