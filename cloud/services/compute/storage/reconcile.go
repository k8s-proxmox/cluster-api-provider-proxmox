package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

const (
	defaultBasePath = "/var/lib/vz"
)

// Reconcile storages used by ProxmoxMachine
func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling storage")
	if err := s.reconcileSnippetStorage(ctx); err != nil {
		log.Error(err, "failed to reconcile snippet storage")
		return err
	}
	if err := s.reconcileImageStorage(ctx); err != nil {
		log.Error(err, "failed to reconcile image storage")
		return err
	}
	log.Info("Reconciled storage")
	return nil
}

// delete storages or keep them if SkipDeletion==true
func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleteing storage")

	// snippet storage
	if !*s.scope.GetStorage().SnippetStorage.SkipDeletion {
		if err := s.deleteStorage(ctx, s.scope.GetStorage().SnippetStorage.Name); err != nil {
			return err
		}
	}

	log.Info("Reconciled storage")
	return nil
}

func (s *Service) reconcileSnippetStorage(ctx context.Context) error {
	opts := generateSnippetStorageOptions(s.scope)
	if err := s.getOrCreateStorage(ctx, opts); err != nil {
		return err
	}
	s.scope.SetSnippetStorage(infrav1.SnippetStorage{Name: opts.Storage, Path: opts.Path})
	return nil
}

// try to get storage
func (s *Service) reconcileImageStorage(ctx context.Context) error {
	name := s.scope.GetStorage().ImageStorage.Name
	if err := s.getStorage(ctx, name, "images", ""); err != nil {
		return err
	}
	s.scope.SetImageStorage(infrav1.ImageStorage{Name: name})
	return nil
}

func (s *Service) getOrCreateStorage(ctx context.Context, opts api.StorageCreateOptions) error {
	log := log.FromContext(ctx)
	if err := s.getStorage(ctx, opts.Storage, opts.Content, opts.StorageType); err != nil {
		if rest.IsNotFound(err) {
			log.Info(fmt.Sprintf("storage %s not found. it will be created", opts.Storage))
			return s.createStorage(ctx, opts)
		}
		log.Error(err, "failed to get storage")
		return err
	}
	return nil
}

// get storage and then confirm if storage meets the requirement
func (s *Service) getStorage(ctx context.Context, name, content, storageType string) error {
	storage, err := s.client.Storage(ctx, name)
	if err != nil {
		return err
	}
	return validateStorage(storage.Storage, content, storageType)
}

// confirm if storage meets the conditions
func validateStorage(storage *api.Storage, content, storageType string) error {
	var err error
	if !strings.Contains(storage.Content, content) {
		err = fmt.Errorf("storage content type is expected to support \"%s\", but supports \"%s\"", content, storage.Content)
	}
	if storageType != "" && storage.Type != storageType {
		err = fmt.Errorf("storage type is expected to be \"%s\", but it's \"%s\": %w", storageType, storage.Type, err)
	}
	return err
}

func (s *Service) createStorage(ctx context.Context, options api.StorageCreateOptions) error {
	if _, err := s.client.CreateStorage(ctx, options.Storage, options.StorageType, options); err != nil {
		return err
	}
	return nil
}

// delete storage
// return error if storage is not empty
func (s *Service) deleteStorage(ctx context.Context, name string) error {
	log := log.FromContext(ctx)
	storage, err := s.client.Storage(ctx, name)
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("storage not found or already deleted")
			return nil
		}
		log.Error(err, "failed to get storage")
		return err
	}

	// check if storage is empty
	storage.Node = s.scope.NodeName()
	contents, err := storage.GetContents(ctx)
	if err != nil {
		log.Error(err, "failed to get content")
		return err
	}
	if len(contents) > 0 {
		return errors.New("Storage must be empty to be deleted")
	}

	// delete
	if err := storage.Delete(ctx); err != nil {
		log.Error(err, "failed to delete storage")
		return err
	}
	return nil
}

// generate storage option for snippet storage
func generateSnippetStorageOptions(scope Scope) api.StorageCreateOptions {
	storageSpec := scope.GetStorage()
	options := api.StorageCreateOptions{
		Storage:     storageSpec.SnippetStorage.Name,
		StorageType: "dir",
		Content:     "snippets",
		Mkdir:       true,
		Path:        storageSpec.SnippetStorage.Path,
	}
	if options.Storage == "" {
		options.Storage = fmt.Sprintf("local-dir-%s", scope.ClusterName())
	}
	if options.Path == "" {
		options.Path = fmt.Sprintf("%s/%s", defaultBasePath, options.Storage)
	}
	return options
}
