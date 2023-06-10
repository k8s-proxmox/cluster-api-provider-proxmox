package storage

import (
	"context"
	"fmt"

	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node/storage"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	vmStorage = "local-capi"
)

func (s *Service) Reconcile(ctx context.Context) error {
	if err := s.createOrGetStorage(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	return nil
}

// createOrGetStorage gets Proxmox Storage for VMs
func (s *Service) createOrGetStorage(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling vm storage")
	_, err := s.client.Storage(vmStorage)
	if err != nil {
		if api.IsNotFound(err) {
			if _, err := s.client.CreateStorage(vmStorage, "dir", defaultVMStorageOptions(vmStorage)); err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

func defaultVMStorageOptions(name string) storage.StorageCreateOptions {
	options := storage.StorageCreateOptions{
		Content: "images,snippets",
		Mkdir:   true,
		Path:    fmt.Sprintf("/var/lib/vz/%s", name),
	}
	return options
}
