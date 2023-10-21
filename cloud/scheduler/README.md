# qemu-scheduler

Scheduling refers to making sure that VM(QEMU) are matched to Proxmox Nodes.

## How to specify vmid
qemu-scheduler reads context and find key registerd to scheduler. If the context has any value of the registerd key, qemu-scheduler uses the plugin that matchies the key.

### Range Plugin
You can specify vmid range with `(start id)-(end id)` format.
```
key: vmid.qemu-scheduler/range
value(example): 100-150
```

### Regex Plugin
```
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
        vmid.qemu-scheduler/range: 100-150 # this annotation will be propagated to your ProxmoxMachine via MachineSet
```