package middleware

import (
	"net/http"

	"context"

	"github.com/remind101/pkg/httpx"
)

// Background is middleware that implements the http.Handler interface to inject
// an initial context object. Use this as the entry point from an http.Handler
// server.
//
// This middleware is deprecated. There is no need to pass a context.Context
// to handlers anymore, since a context object is available from the request. Once
// we update the signature for httpx.Handler to remove the context parameter, this
// middleware can be removed.
type Background struct {
	// The wrapped httpx.Handler to call down to.
	handler httpx.Handler
}

func BackgroundContext(h httpx.Handler) *Background {
	return &Background{
		handler: h,
	}
}

// ServeHTTP implements the http.Handler interface.
func (h *Background) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTPContext(r.Context(), w, r)
}

func (h *Background) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.handler.ServeHTTPContext(ctx, w, r)
}
