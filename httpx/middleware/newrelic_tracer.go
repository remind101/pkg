package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/remind101/newrelic"
)

// newTxFromMuxRoute returns a new newrelic.Tx from a gorilla mux route.
func newTxFromMuxRoute(tracer newrelic.TxTracer) func(*http.Request) newrelic.Tx {
	return func(r *http.Request) newrelic.Tx {
		path := r.URL.Path

		if route := mux.CurrentRoute(r); route != nil {
			path, _ = route.GetPathTemplate()
		}

		txName := fmt.Sprintf("%s %s", r.Method, path)

		t := newrelic.NewRequestTx(txName, r.URL.String())
		t.Tracer = tracer
		return t
	}
}

// NewRelicTracer is middleware that can be wrapped around a mux.Route to add
// NewRelic transaction traces.
type NewRelicTracer struct {
	// newTx will be called to create a new newrelic.Tx for the
	// request.
	newTx func(*http.Request) newrelic.Tx

	handler http.Handler
}

// NewRelicTracing wraps h with NewRelic tracing for gorilla mux routes. It's
// important that this middleware is added AFTER the route has been handled by
// mux. In other words:
//
// BAD
//	NewRelicTracing(mux.NewRouter(), tracer)
//
// GOOD
//	router.Handle("/users/{id}", NewRelicTracing(handler, tracer))
func NewRelicTracing(h http.Handler, tracer newrelic.TxTracer) *NewRelicTracer {
	return &NewRelicTracer{
		newTx:   newTxFromMuxRoute(tracer),
		handler: h,
	}
}

func (h *NewRelicTracer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tx := h.newTx(r)
	tx.Start()
	defer tx.End()

	r = r.WithContext(newrelic.WithTx(r.Context(), tx))
	h.handler.ServeHTTP(w, r)
}
