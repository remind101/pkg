package middleware

import (
	"fmt"
	"net/http"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
)

// Recovery is a middleware that will recover from panics and return the error.
type Recovery struct {
	// Reporter is a Reporter that will be inserted into the context. It
	// will also be used to report panics.
	reporter.Reporter

	// handler is the wrapped httpx.Handler.
	handler http.Handler
}

func Recover(h http.Handler, r reporter.Reporter) *Recovery {
	return &Recovery{
		Reporter: r,
		handler:  h,
	}
}

// ServeHTTP implements the http.Handler interface. It recovers from
// panics and returns an error for upstream middleware to handle.
func (h *Recovery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := reporter.WithReporter(r.Context(), h.Reporter)

	// Add the request to the context.
	reporter.AddRequest(ctx, r)

	// Add the request id
	reporter.AddContext(ctx, "request_id", httpx.RequestID(ctx))

	defer func() {
		if v := recover(); v != nil {
			err := fmt.Errorf("%v", v)

			if v, ok := v.(error); ok {
				err = v
			}

			reporter.Report(ctx, err)
		}
	}()

	h.handler.ServeHTTP(w, r)
}
