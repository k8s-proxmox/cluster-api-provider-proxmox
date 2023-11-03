package instance

import "context"

// add vm to resource pool
func (s *Service) reconcileResourcePool(ctx context.Context) error {
	pool, err := s.client.Pool(ctx, s.scope.GetResourcePool())
	if err != nil {
		return err
	}
	return pool.AddVMs(ctx, []int{*s.scope.GetVMID()})
}
