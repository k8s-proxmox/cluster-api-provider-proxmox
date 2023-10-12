package storage

import (
	"context"
	"errors"
	"fmt"
	"k8s.io/utils/pointer"
	"strings"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
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
			log.Info(fmt.Sprintf("storage %s not found. it will be created", opts.Storage))
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

	if options.Mkdir != nil && *options.Mkdir == false {
		if err := s.createStorageDirs(ctx, options); err != nil {
			return err
		}
	}

	if _, err := s.client.CreateStorage(ctx, options.Storage, options.StorageType, options); err != nil {
		return err
	}

	return nil
}

func (s *Service) createStorageDirs(ctx context.Context, options api.StorageCreateOptions) error {
	log := log.FromContext(ctx)

	nodes, err := s.client.Nodes(ctx)
	if err != nil {
		return err
	}

	hasFailure := false
	for _, node := range nodes {
		vnc, err := s.client.NewNodeVNCWebSocketConnection(context.TODO(), node.Node)
		if err != nil {
			log.Error(err, fmt.Sprintf("Failed to create shell to node %s", node.Node))
			hasFailure = true
		}

		for _, dir := range strings.Split(options.Content, ",") {
			_, _, err := vnc.Exec(ctx, fmt.Sprintf("mkdir -p %s/%s", options.Path, dir))
			if err != nil {
				log.Error(err, fmt.Sprintf("Failed creating content dir %s", dir))
				hasFailure = true
			}
		}
		vnc.Close()
	}

	if hasFailure {
		return fmt.Errorf("failed creating content directories for storage %s", options.Storage)
	} else {
		return nil
	}
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
			return errors.New("Storage must be empty to be deleted")
		}
	}

	// delete
	if err := storage.Delete(ctx); err != nil {
		return err
	}
	return nil
}

func generateVMStorageOptions(scope Scope) api.StorageCreateOptions {

	// when using non-root user, we need to create directories ourselves. Otherwise they will be owned by root.
	// We assume we have permission for base storage path.
	mkdirs := scope.CloudClient().RESTClient().Credentials().Username == "root@pam"

	storageSpec := scope.Storage()
	options := api.StorageCreateOptions{
		Storage:     storageSpec.Name,
		StorageType: "dir",
		Content:     "images,snippets",
		Mkdir:       pointer.Bool(mkdirs),
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
