package svc

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/reporter"
)

type HTTPHandlerOpts struct {
	Reporter          reporter.Reporter
	ForwardingHeaders []string
	BasicAuth         string
	ErrorHandler      middleware.ErrorHandlerFunc
	HandlerTimeout    time.Duration
}

func NewHTTPStack(h http.Handler, opts HTTPHandlerOpts) http.Handler {
	var hx httpx.Handler

	// Adapter for httpx middlewares. No middleware will rely on the error return value
	// or context arguement.
	hx = httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(rw, r)
		return nil
	})

	// Recover from panics. A panic is converted to an error. This should be first,
	// even though it means panics in middleware will not be recovered, because
	// later middleware expects endpoint panics to be returned as an error.
	hx = middleware.BasicRecover(hx)

	// Handler errors returned by endpoint handler or recovery middleware.
	// Errors will no longer be returned after this middeware.
	errorHandler := opts.ErrorHandler
	if errorHandler == nil {
		errorHandler = middleware.ReportingErrorHandler
	}
	hx = middleware.HandleError(hx, errorHandler)

	// Insert logger into context and log requests at INFO level.
	hx = middleware.LogTo(hx, middleware.LoggerWithRequestID)

	// Add reporter to context and request to reporter context.
	hx = middleware.WithReporter(hx, opts.Reporter)

	// Add the request id to the context.
	hx = middleware.ExtractRequestID(hx)

	// Add basic auth
	if opts.BasicAuth != "" {
		user := strings.Split(opts.BasicAuth, ":")[0]
		pass := strings.Split(opts.BasicAuth, ":")[1]
		hx = middleware.BasicAuth(hx, user, pass, "")
	}

	// Adds forwarding headers from request to the context. This allows http clients
	// to get those headers from the context and add them to upstream requests.
	if len(opts.ForwardingHeaders) > 0 {
		for _, header := range opts.ForwardingHeaders {
			hx = middleware.ExtractHeader(hx, header)
		}
	}

	// Wrap the route in middleware to add a context.Context. This middleware must be
	// last as it acts as the adaptor between http.Handler and httpx.Handler.
	return middleware.BackgroundContext(hx)
}
