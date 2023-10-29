package framework

import "context"

type CtxKey string

// bind map's key-value to context key-value.
// type of key is translated to CtxKey type
func ContextWithMap(ctx context.Context, m map[string]string) context.Context {
	for key, value := range m {
		ctx = context.WithValue(ctx, CtxKey(key), value)
	}
	return ctx
}
