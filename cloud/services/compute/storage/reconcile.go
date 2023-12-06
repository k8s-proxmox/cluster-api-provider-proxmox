package storage

import (
	"context"

	"github.com/k8s-proxmox/proxmox-go/api"
	"github.com/k8s-proxmox/proxmox-go/proxmox"
	"github.com/k8s-proxmox/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/k8s-proxmox/cluster-api-provider-proxmox/api/v1beta1"
)

const (
	DefaultBasePath = "/var/lib/vz"
)

func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling storage")

	if err := s.createOrGetStorage(ctx); err != nil {
		return err
	}

	log.Info("Reconciled storage")
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleteing storage")
	return s.deleteStorage(ctx)
}

// createOrGetStorage gets Proxmox Storage for VMs
func (s *Service) createOrGetStorage(ctx context.Context) error {
	log := log.FromContext(ctx)
	opts := generateVMStorageOptions(s.scope)
	if err := s.getStorage(ctx, opts.Storage); err != nil {
		if rest.IsNotFound(err) {
			log.Info("storage %s not found. it will be created")
			return s.createStorage(ctx, opts)
		}
		return err
	}

	s.scope.SetStorage(infrav1.Storage{Name: opts.Storage, Path: opts.Path})
	return nil
}

func (s *Service) getStorage(ctx context.Context, name string) error {
	if _, err := s.client.Storage(ctx, name); err != nil {
		return err
	}
	return nil
}

func (s *Service) createStorage(ctx context.Context, options api.StorageCreateOptions) error {
	if _, err := s.client.CreateStorage(ctx, options.Storage, options.StorageType, options); err != nil {
		return err
	}
	return nil
}

func (s *Service) deleteStorage(ctx context.Context) error {
	log := log.FromContext(ctx)

	var storage *proxmox.Storage
	storage, err := s.client.Storage(ctx, s.scope.Storage().Name)
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("storage not found or already deleted")
			return nil
		}
		return err
	}

	nodes, err := s.client.Nodes(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		storage.Node = node.Node

		// check if storage is empty
		contents, err := storage.GetContents(ctx)
		if err != nil {
			return err
		}
		if len(contents) > 0 {
			log.Info("storage not empty, skipping deletion")
			return nil
		}
	}

	// delete
	if err := storage.Delete(ctx); err != nil {
		return err
	}
	return nil
}

func generateVMStorageOptions(scope Scope) api.StorageCreateOptions {
	storageSpec := scope.Storage()
	mkdir := true
	options := api.StorageCreateOptions{
		Storage:     storageSpec.Name,
		StorageType: "dir",
		Content:     "snippets",
		Mkdir:       &mkdir,
		Path:        storageSpec.Path,
	}
	return options
}
