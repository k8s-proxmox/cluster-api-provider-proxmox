package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

const (
	defaultBasePath = "/var/lib/vz"
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
	return s.deleteStorage(ctx)
}

// createOrGetStorage gets Proxmox Storage for VMs
func (s *Service) createOrGetStorage(ctx context.Context) error {
	opts := generateVMStorageOptions(s.scope)
	if err := s.getStorage(ctx, opts.Storage); err != nil {
		if rest.IsNotFound(err) {
			if err := s.createStorage(ctx, opts); err != nil {
				return err
			}
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
	nodes, err := s.client.Nodes(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		storage, err := s.client.Storage(ctx, s.scope.Storage().Name)
		if err != nil {
			log.Info(err.Error())
			continue
		}
		storage.Node = node.Node

		// check if storage is empty
		contents, err := storage.GetContents(ctx)
		if err != nil {
			return err
		}
		if len(contents) > 0 {
			return errors.New("Storage must be empty to be deleted")
		}

		// delete
		if err := storage.Delete(ctx); err != nil {
			log.Info(err.Error())
			return err
		}
	}
	return nil
}

func generateVMStorageOptions(scope Scope) api.StorageCreateOptions {
	storageSpec := scope.Storage()
	options := api.StorageCreateOptions{
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
