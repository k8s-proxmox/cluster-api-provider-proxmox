package resourcepool

import (
	"context"
	"fmt"

	"github.com/sp-yduck/proxmox-go/api"
	"github.com/sp-yduck/proxmox-go/proxmox"
	"github.com/sp-yduck/proxmox-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	DefaultBasePath = "/var/lib/vz"
)

func (s *Service) Reconcile(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Reconciling resource pool")

	pool, err := s.createOrGetResourcePool(ctx)
	if err != nil {
		return err
	}

	if err := pool.AddStorages(ctx, []string{s.scope.Storage().Name}); err != nil {
		log.Error(err, "failed to add sotrage to pool")
		return err
	}

	log.Info("Reconciled resource pool")
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	log := log.FromContext(ctx)
	log.Info("Deleteing resource pool")
	return s.deleteResourcePool(ctx)
}

func (s *Service) createOrGetResourcePool(ctx context.Context) (*proxmox.Pool, error) {
	log := log.FromContext(ctx)

	pool, err := s.client.Pool(ctx, s.scope.ResourcePool())
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("resource pool not found. it will be created")
			return s.createResourcePool(ctx)
		}
		log.Error(err, "failed to get resource pool")
		return nil, err
	}
	return pool, nil
}

func (s *Service) createResourcePool(ctx context.Context) (*proxmox.Pool, error) {
	pool := api.ResourcePool{
		PoolID:  s.scope.ResourcePool(),
		Comment: fmt.Sprintf("Default Resource Pool used for %s cluster", s.scope.Name()),
	}
	return s.client.CreatePool(ctx, pool)
}

func (s *Service) deleteResourcePool(ctx context.Context) error {
	log := log.FromContext(ctx)
	poolid := s.scope.ResourcePool()
	pool, err := s.client.Pool(ctx, poolid)
	if err != nil {
		if rest.IsNotFound(err) {
			log.Info("resource pool not found or already deleted")
			return nil
		}
		return err
	}
	members, err := pool.GetMembers(ctx)
	if err != nil {
		return err
	}
	if len(members) != 0 {
		log.Info("resource pool not empty, skipping deletion")
		return nil
	}
	return s.client.DeletePool(ctx, poolid)
}
