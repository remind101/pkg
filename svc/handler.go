// Package svc provides some tooling to make building services with remind101/pkg
// easier.
//
// Recommend Usage:
//
//	func main() {
//		env := svc.InitAll()
//		defer env.Close()
//
//		r := httpx.NewRouter()
//		// ... add routes
//
//		h := svc.NewStandardHandler(svc.HandlerOpts{
//			Router:   r,
//			Reporter: env.Reporter,
//	})
//
// 	s := svc.NewServer(h, svc.WithPort("8080"))
//  svc.RunServer(s)
// }
package svc

import (
	"net/http"
	"strings"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/reporter"
)

type HandlerOpts struct {
	Router            httpx.Handler
	Reporter          reporter.Reporter
	ForwardingHeaders []string
	BasicAuth         string
	ErrorHandler      middleware.ErrorHandlerFunc
	HandlerTimeout    time.Duration
}

// NewStandardHandler returns an http.Handler with a standard middleware stack.
// The last middleware added is the first middleware to handle the request.
// Order is pretty important as some middleware depends on others having run
// already.
func NewStandardHandler(opts HandlerOpts) http.Handler {
	h := opts.Router

	if opts.HandlerTimeout != 0 {
		// Timeout requests after the given Timeout duration.
		h = middleware.TimeoutHandler(h, opts.HandlerTimeout)
	}

	// Recover from panics. A panic is converted to an error. This should be first,
	// even though it means panics in middleware will not be recovered, because
	// later middleware expects endpoint panics to be returned as an error.
	h = middleware.BasicRecover(h)

	// Handler errors returned by endpoint handler or recovery middleware.
	// Errors will no longer be returned after this middeware.
	errorHandler := opts.ErrorHandler
	if errorHandler == nil {
		errorHandler = middleware.ReportingErrorHandler
	}
	h = middleware.HandleError(h, errorHandler)

	// Add request tracing. Must go after the HandleError middleware in order
	// to capture the status code written to the response.
	h = middleware.OpentracingTracing(h)

	// Insert logger into context and log requests at INFO level.
	h = middleware.LogTo(h, middleware.LoggerWithRequestID)

	// Add reporter to context and request to reporter context.
	h = middleware.WithReporter(h, opts.Reporter)

	// Add the request id to the context.
	h = middleware.ExtractRequestID(h)

	// Add basic auth
	if opts.BasicAuth != "" {
		user := strings.Split(opts.BasicAuth, ":")[0]
		pass := strings.Split(opts.BasicAuth, ":")[1]
		h = middleware.BasicAuth(h, user, pass, "")
	}

	// Adds forwarding headers from request to the context. This allows http clients
	// to get those headers from the context and add them to upstream requests.
	if len(opts.ForwardingHeaders) > 0 {
		for _, header := range opts.ForwardingHeaders {
			h = middleware.ExtractHeader(h, header)
		}
	}

	// Wrap the route in middleware to add a context.Context. This middleware must be
	// last as it acts as the adaptor between http.Handler and httpx.Handler.
	return middleware.BackgroundContext(h)
}
