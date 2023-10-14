package fake

import (
	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"k8s.io/utils/pointer"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

type FakeMachineScope struct {
	cloudClient    *proxmox.Service
	clusterName    string
	name           string
	namespace      string
	nodeName       string
	biosUUID       string
	image          infrav1.Image
	providerID     string
	bootstrapData  string
	instanceStatus *infrav1.InstanceStatus
	storage        infrav1.Storage
	cloudInit      infrav1.CloudInit
	network        infrav1.Network
	hardware       infrav1.Hardware
	vmid           *int
	options        infrav1.Options
}

func NewMachineScope(client *proxmox.Service) *FakeMachineScope {
	return &FakeMachineScope{
		cloudClient: client,
		clusterName: "foo-cluster",
		name:        "foo-machine",
		namespace:   "default",
		image: infrav1.Image{
			Checksum:     "c5eed826009c9f671bc5f7c9d5d63861aa2afe91aeff1c0d3a4cb5b28b2e35d6",
			ChecksumType: pointer.String("sha256"),
			URL:          "https://cloud-images.ubuntu.com/releases/jammy/release-20230914/ubuntu-22.04-server-cloudimg-amd64-disk-kvm.img",
		},
		bootstrapData: `#cloud-config
write_files:
- path: /tmp/cappx-test
	owner: root:root
	permissions: '0640'
	content: |
		test
`,
		hardware: infrav1.Hardware{
			Memory: 2048,
			CPU:    1,
			Disk:   "50G",
		},
	}
}

func (f *FakeMachineScope) CloudClient() *proxmox.Service {
	return f.cloudClient
}

func (f *FakeMachineScope) ClusterName() string {
	return f.clusterName
}

func (f *FakeMachineScope) Name() string {
	return f.name
}

func (f *FakeMachineScope) Namespace() string {
	return f.namespace
}

func (f *FakeMachineScope) NodeName() string {
	return f.nodeName
}

func (f *FakeMachineScope) GetBiosUUID() *string {
	return &f.biosUUID
}

func (f *FakeMachineScope) GetImage() infrav1.Image {
	return f.image
}

func (f *FakeMachineScope) GetProviderID() string {
	return f.providerID
}

func (f *FakeMachineScope) GetBootstrapData() (string, error) {
	return f.bootstrapData, nil
}

func (f *FakeMachineScope) GetInstanceStatus() *infrav1.InstanceStatus {
	return f.instanceStatus
}

func (f *FakeMachineScope) GetStorage() infrav1.Storage {
	f.storage.SnippetStorage.SkipDeletion = pointer.Bool(false)
	return f.storage
}

func (f *FakeMachineScope) GetCloudInit() infrav1.CloudInit {
	return f.cloudInit
}

func (f *FakeMachineScope) GetNetwork() infrav1.Network {
	return f.network
}

func (f *FakeMachineScope) GetHardware() infrav1.Hardware {
	return f.hardware
}

func (f *FakeMachineScope) GetVMID() *int {
	return f.vmid
}

func (f *FakeMachineScope) GetOptions() infrav1.Options {
	return f.options
}

func (f *FakeMachineScope) SetName(name string) {
	f.name = name
}

func (f *FakeMachineScope) SetNamespace(namespace string) {
	f.namespace = namespace
}

func (f *FakeMachineScope) SetProviderID(uuid string) error {
	f.biosUUID = uuid
	f.providerID = "proxmox://" + uuid
	return nil
}

func (f *FakeMachineScope) SetInstanceStatus(status infrav1.InstanceStatus) {
	f.instanceStatus = &status
}

func (f *FakeMachineScope) SetNodeName(name string) {
	f.nodeName = name
}

func (f *FakeMachineScope) SetVMID(vmid int) {
	f.vmid = &vmid
}

func (f *FakeMachineScope) SetSnippetStorage(storage infrav1.SnippetStorage) {
	f.storage.SnippetStorage = storage
}

func (f *FakeMachineScope) SetImageStorage(storage infrav1.ImageStorage) {
	f.storage.ImageStorage = storage
}

func (f *FakeMachineScope) SetConfigStatus(config api.VirtualMachineConfig) {
}

func (f *FakeMachineScope) PatchObject() error {
	return nil
}
