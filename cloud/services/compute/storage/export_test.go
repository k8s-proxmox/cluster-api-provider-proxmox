package storage

import (
	"github.com/sp-yduck/proxmox-go/api"
)

func GenerateVMStorageOptions(scope Scope) api.StorageCreateOptions {
	return generateVMStorageOptions(scope)
}
