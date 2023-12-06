package storage

import (
	"github.com/k8s-proxmox/proxmox-go/api"
)

func GenerateVMStorageOptions(scope Scope) api.StorageCreateOptions {
	return generateVMStorageOptions(scope)
}
