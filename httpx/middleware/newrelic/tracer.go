package newrelic

import (
	"fmt"
	"net/http"

	"context"

	"github.com/newrelic/go-agent"
	"github.com/remind101/pkg/httpx"
)

type NewRelicGoTracer struct {
	handler httpx.Handler
	app     newrelic.Application
	router  *httpx.Router
}

func NewRelicGoTracing(h httpx.Handler, router *httpx.Router, app newrelic.Application) httpx.Handler {
	if app != nil {
		return &NewRelicGoTracer{h, app, router}
	} else {
		return h
	}
}

func (h *NewRelicGoTracer) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	path := templatePath(h.router, r)
	txName := fmt.Sprintf("%s %s", r.Method, path)

	txn := h.app.StartTransaction(txName, w, r)

	ctx = context.WithValue(ctx, newrelic_txn, txn)
	ctx = context.WithValue(ctx, newrelic_app, h.app)

	defer txn.End()

	return h.handler.ServeHTTPContext(ctx, w, r)
}

func NewrelicAppFromContext(ctx context.Context) (newrelic.Application, bool) {
	app, ok := ctx.Value(newrelic_app).(newrelic.Application)
	return app, ok
}

func NewrelicTxnFromContext(ctx context.Context) (newrelic.Transaction, bool) {
	txn, ok := ctx.Value(newrelic_txn).(newrelic.Transaction)
	return txn, ok
}

const (
	newrelic_app = iota
	newrelic_txn = iota
)

func templatePath(router *httpx.Router, r *http.Request) string {
	var tpl string

	route, _, _ := router.Handler(r)
	if route != nil {
		tpl = route.GetPathTemplate()
	}

	if tpl == "" {
		tpl = r.URL.Path
	}

	return tpl
}
