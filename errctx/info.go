package errctx

import (
	"context"
	"net/http"
)

// info is used internally to store contextual information.
type info struct {
	data    map[string]interface{}
	request *http.Request
}

func newInfo() *info {
	return &info{data: make(map[string]interface{})}
}

func withInfo(ctx context.Context) context.Context {
	if _, ok := infoFromContext(ctx); ok {
		return ctx
	}
	return context.WithValue(ctx, infoKey, newInfo())
}

func infoFromContext(ctx context.Context) (*info, bool) {
	i, ok := ctx.Value(infoKey).(*info)
	return i, ok
}

// key used to store context values from within this package.
type key int

const (
	infoKey key = iota
)
