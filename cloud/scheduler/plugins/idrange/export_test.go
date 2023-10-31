package idrange

import "context"

func FindVMIDRange(ctx context.Context) (int, int, error) {
	return findVMIDRange(ctx)
}
