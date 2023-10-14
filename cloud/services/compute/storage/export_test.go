package storage

import (
	"github.com/sp-yduck/proxmox-go/api"
)

func ValidateStorage(storage *api.Storage, content, storageType string) error {
	return validateStorage(storage, content, storageType)
}

func GenerateSnippetStorageOptions(scope Scope) api.StorageCreateOptions {
	return generateSnippetStorageOptions(scope)
}
