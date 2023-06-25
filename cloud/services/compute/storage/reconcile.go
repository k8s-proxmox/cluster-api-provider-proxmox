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

	opts := generateVMStorageOptions(s.scope)
	if err := s.getStorage(opts.Storage); err != nil {
		if api.IsNotFound(err) {
			if err := s.createStorage(opts); err != nil {
				return err
			}
		}
		return err
	}

	s.scope.SetStorage(infrav1.Storage{Name: opts.Storage, Path: opts.Path})
	return nil
}

func (s *Service) getStorage(name string) error {
	if _, err := s.client.Storage(name); err != nil {
		return err
	}
	return nil
}

func (s *Service) createStorage(options storage.StorageCreateOptions) error {
	if _, err := s.client.CreateStorage(options.Storage, options.StorageType, options); err != nil {
		return err
	}
	return nil
}

func generateVMStorageOptions(scope Scope) storage.StorageCreateOptions {
	storageSpec := scope.Storage()
	options := storage.StorageCreateOptions{
		Storage:     storageSpec.Name,
		StorageType: "dir",
		Content:     "images,snippets",
		Mkdir:       true,
		Path:        storageSpec.Path,
	}
	if options.Storage == "" {
		options.Storage = fmt.Sprintf("local-dir-%s", scope.Name())
	}
	if options.Path == "" {
		options.Path = fmt.Sprintf("%s/%s", defaultBasePath, options.Storage)
	}
	return options
}
