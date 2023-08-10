package v1beta1

import "strconv"

// import "encoding/json"

// +kubebuilder:validation:Enum:=x86_64;aarch64
type Arch string

// +kubebuilder:validation:Enum:=seabios;ovmf
type BIOS string

// +kubebuilder:validation:Enum:=0;2;1024
type HugePages int

// +kubebuilder:validation:Enum:=backup;clone;create;migrate;rollback;snapshot;snapshot-delete;suspending;suspended
type Lock string

// +kubebuilder:validation:Enum:=other;wxp;w2k;w2k3;w2k8;wvista;win7;win8;win10;win11;l24;l26;solaris
type OSType string

// +kubebuilder:validation:Pattern:="[a-zA-Z0-9-_.;]+"
type Tag string

type Tags []Tag

func (h *HugePages) String() string {
	if h == nil {
		return ""
	} else if *h == 0 {
		return "any"
	}
	return strconv.Itoa(int(*h))
}

func (t *Tags) String() string {
	var tags string
	for _, tag := range *t {
		tags += string(tag) + ";"
	}
	return tags
}

// Options
type Options struct {
	// Enable/Disable ACPI. Defaults to true.
	ACPI bool `json:"acpi,omitempty"`

	// Virtual processor architecture. Defaults to the host. x86_64 or aarch64.
	Arch Arch `json:"arch,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	// Amount of target RAM for the VM in MiB. Using zero disables the ballon driver.
	Balloon int `json:"balloon,omitempty"`

	// Description for the VM. Shown in the web-interface VM's summary.
	// This is saved as comment inside the configuration file.
	Description string `json:"description,omitempty"`

	// Script that will be executed during various steps in the vms lifetime.
	// HookScripts []Hookscript `json:"hookScripts,omitempty"`

	// enable hotplug feature. list og devices.
	// network, disk, cpu, memory, usb. Defaults to [network, disk, usb].
	// HotPlug []HotPlugDevice `json:"hotPlug,omitempty"`

	// enable/disable hugepages memory. 0 or 2 or 1024. 0 indicated 'any'
	HugePages *HugePages `json:"hugePages,omitempty"`

	// Use together with hugepages. If enabled, hugepages will not not be deleted
	// after VM shutdown and can be used for subsequent starts. Defaults to false.
	KeepHugePages bool `json:"keepHugePages,omitempty"`

	// Enable/disable KVM hardware virtualization. Defaults to true.
	KVM bool `json:"kvm,omitempty"`

	// Set the real time clock (RTC) to local time.
	// This is enabled by default if the `ostype` indicates a Microsoft Windows OS.
	LocalTime bool `json:"localTime,omitempty"`

	// Lock/unlock the VM.
	Lock Lock `json:"lock,omitempty"`

	// Set maximum tolerated downtime (in seconds) for migrations.
	// MigrateDowntime json.Number `json:"migrateDowntime,omitempty"`

	// Set maximum speed (in MB/s) for migrations. Value 0 is no limit.
	// MigrateSpeed `json:"migrateSpeed,omitempty"`

	// Enable/disable NUMA.
	NUMA bool `json:"numa,omitempty"`

	// Specifies whether a VM will be started during system bootup.
	OnBoot bool `json:"onBoot,omitempty"`

	// Specify guest operating system. This is used to enable special
	// optimization/features for specific operating systems.
	OSType OSType `json:"osType,omitempty"`

	// Sets the protection flag of the VM.
	// This will disable the remove VM and remove disk operations.
	// Defaults to false.
	Protection bool `json:"protection,omitempty"`

	// Allow reboot. If set to 'false' the VM exit on reboot.
	// Defaults to true.
	Reboot bool `json:"reboot,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:validation:Maximum:=5000
	// Amount of memory shares for auto-ballooning. The larger the number is, the more memory this VM gets.
	// Number is relative to weights of all other running VMs. Using zero disables auto-ballooning.
	// Auto-ballooning is done by pvestatd. 0 ~ 5000. Defaults to 1000.
	Shares int `json:"shares,omitempty"`

	// Set the initial date of the real time clock.
	// Valid format for date are:'now' or '2006-06-17T16:01:21' or '2006-06-17'.
	// Defaults to 'now'.
	// StartDate string `json:"startDate,omitempty"`

	// StartUp string `json:"startUp,omitempty`

	// Enable/disable the USB tablet device. This device is usually needed to allow
	// absolute mouse positioning with VNC. Else the mouse runs out of sync with normal VNC clients.
	// If you're running lots of console-only guests on one host,
	// you may consider disabling this to save some context switches.
	// This is turned off by default if you use spice (`qm set <vmid> --vga qxl`).
	// Defaults to true.
	Tablet bool `json:"tablet,omitempty"`

	// Tags of the VM. This is only meta information.
	Tags Tags `json:"tags,omitempty"`

	// Enable/disable time drift fix. Defaults to false.
	TimeDriftFix bool `json:"timeDriftFix,omitempty"`

	// Enable/disable Template. Defaults to false.
	Template bool `json:"template,omitempty"`

	// TPMState string `json:"tpmState,omitempty"`

	// +kubebuilder:validation:Minimum:=0
	// Number of hotplugged vcpus. Defaults to 0.
	VCPUs int `json:"vcpus,omitempty"`

	// VGA string `json:"vga,omitempty"`

	// +kubebuilder:validation:Pattern:="(?:[a-fA-F0-9]{8}(?:-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}|[01])"
	// The VM generation ID (vmgenid) device exposes a 128-bit integer value identifier to the guest OS.
	// This allows to notify the guest operating system when the virtual machine is executed with a different configuration
	// (e.g. snapshot execution or creation from a template).
	// The guest operating system notices the change, and is then able to react as appropriate by marking its copies of distributed databases as dirty,
	// re-initializing its random number generator, etc.
	// Note that auto-creation only works when done through API/CLI create or update methods, but not when manually editing the config file.
	// regex: (?:[a-fA-F0-9]{8}(?:-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}|[01]). Defaults to 1 (autogenerated)
	VMGenerationID string `json:"vmGenerationID,omitempty"`

	// Default storage for VM state volumes/files.
	// VMStateStorage string `json:"vmStateStorage,omitempty"`

	// Create a virtual hardware watchdog device. Once enabled (by a guest action),
	// the watchdog must be periodically polled by an agent inside the guest or else
	// the watchdog will reset the guest (or execute the respective action specified)
	// WatchDog string `json:"watchDog,omitempty"`
}
