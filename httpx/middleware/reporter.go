package middleware

import (
	"context"
	"net/http"

	"github.com/remind101/pkg/errctx"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
)

// Reporter is a middleware that adds a Reporter to the request context and adds
// the request info to the reporter context.
type Reporter struct {
	handler  httpx.Handler
	reporter reporter.Reporter
}

func (m *Reporter) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Add reporter to context.
	ctx = reporter.WithReporter(ctx, m.reporter)

	// Add the request to the reporter context.
	ctx = errctx.WithRequest(ctx, r)

	// Add the request id to reporter context.
	ctx = errctx.WithInfo(ctx, "request_id", httpx.RequestID(ctx))

	return m.handler.ServeHTTPContext(ctx, w, r)
}

func WithReporter(h httpx.Handler, r reporter.Reporter) *Reporter {
	return &Reporter{handler: h, reporter: r}
}
