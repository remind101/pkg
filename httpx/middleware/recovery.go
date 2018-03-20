package middleware

import (
	"fmt"
	"net/http"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
)

// Recovery is a middleware that will recover from panics and return the error.
type Recovery struct {
	// Reporter is a Reporter that will be inserted into the context. It
	// will also be used to report panics.
	reporter.Reporter

	// handler is the wrapped httpx.Handler.
	handler httpx.Handler
}

func Recover(h httpx.Handler, r reporter.Reporter) *Recovery {
	return &Recovery{
		Reporter: r,
		handler:  h,
	}
}

// ServeHTTPContext implements the httpx.Handler interface. It recovers from
// panics and returns an error for upstream middleware to handle.
func (h *Recovery) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	ctx = reporter.WithReporter(ctx, h.Reporter)

	// Add the request to the context.
	ctx = reporter.AddRequest(ctx, r)

	// Add the request id
	ctx = reporter.AddContext(ctx, "request_id", httpx.RequestID(ctx))

	defer func() {
		if v := recover(); v != nil {
			w.WriteHeader(http.StatusInternalServerError)

			var ok bool
			if err, ok = v.(error); !ok {
				err = fmt.Errorf("%v", v)
			}

			reporter.Report(ctx, err)
			return
		}
	}()

	err = h.handler.ServeHTTPContext(ctx, w, r)

	return
}

type BasicRecovery struct {
	handler httpx.Handler
}

// ServeHTTPContext implements the httpx.Handler interface. It recovers from
// panics and returns an error for upstream middleware to handle.
func (h *BasicRecovery) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if v := recover(); v != nil {
			var ok bool
			if err, ok = v.(error); !ok {
				err = fmt.Errorf("%v", v)
			}
			return
		}
	}()

	err = h.handler.ServeHTTPContext(ctx, w, r)

	return
}

func BasicRecover(h httpx.Handler) *BasicRecovery {
	return &BasicRecovery{handler: h}
}
