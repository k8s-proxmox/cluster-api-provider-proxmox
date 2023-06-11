package storage

import (
	"context"
	"fmt"

	"github.com/sp-yduck/proxmox/pkg/api"
	"github.com/sp-yduck/proxmox/pkg/service/node/storage"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

const (
	defaultBasePath = "/var/lib/vz"
)

func (s *Service) Reconcile(ctx context.Context) error {
	if err := s.createOrGetStorage(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	// to do
	return nil
}

// createOrGetStorage gets Proxmox Storage for VMs
func (s *Service) createOrGetStorage(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling vm storage")
	if err := s.getStorage(); err != nil {
		if api.IsNotFound(err) {
			if err := s.createStorage(); err != nil {
				return err
			}
		}
		return err
	}
	return nil
}

func (s *Service) getStorage() error {
	storageSpec := s.scope.Storage()
	if _, err := s.client.Storage(storageSpec.Name); err != nil {
		return err
	}
	return nil
}

func (s *Service) createStorage() error {
	storageSpec := s.scope.Storage()
	opts := storage.StorageCreateOptions{
		Content: "images,snippets",
		Mkdir:   true,
		Path:    generateStoragePath(storageSpec),
	}
	if _, err := s.client.CreateStorage(storageSpec.Name, "dif", opts); err != nil {
		return err
	}
	return nil
}

func generateStoragePath(storage infrav1.Storage) string {
	if storage.Path == "" {
		return fmt.Sprintf("%s/%s", defaultBasePath, storage.Name)
	}
	return storage.Path
}

func defaultVMStorageOptions(name string) storage.StorageCreateOptions {
	options := storage.StorageCreateOptions{
		Content: "images,snippets",
		Mkdir:   true,
		Path:    defaultBasePath + "/" + name,
	}
	return options
}
