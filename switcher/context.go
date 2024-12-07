package switcher

import "context"

type nameCtxKeyType struct{}

var nameCtxKey = nameCtxKeyType{}

func WithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, nameCtxKey, name)
}

func GetName(ctx context.Context) string {
	v := ctx.Value(nameCtxKey)
	if v == nil {
		return ""
	}
	return v.(string)
}
