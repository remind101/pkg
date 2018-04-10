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
// 	s := svc.NewServer(h, WithPort("8080"))
//  svc.RunServer(s)
// }
package svc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	ddtrace "github.com/DataDog/dd-trace-go/opentracing"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/metrics"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/reporter/rollbar"
)

type HandlerOpts struct {
	Router            *httpx.Router
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
	h := httpx.Handler(opts.Router)

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
	h = middleware.OpentracingTracing(h, opts.Router)

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

// NewServerOpt allows users to customize the http.Server used by RunServer.
type NewServerOpt func(*http.Server)

// ServerDefaults specifies default server options to use for RunServer.
var ServerDefaults = func(srv *http.Server) {
	srv.Addr = ":8080"
	srv.WriteTimeout = 5 * time.Second
	srv.ReadHeaderTimeout = 5 * time.Second
	srv.IdleTimeout = 120 * time.Second
}

// WithPort sets the port for the server to run on.
func WithPort(port string) NewServerOpt {
	return func(srv *http.Server) {
		srv.Addr = ":" + port
	}
}

// NewServer offers some convenience and good defaults for creating an http.Server
func NewServer(h http.Handler, opts ...NewServerOpt) *http.Server {
	srv := &http.Server{Handler: h}

	// Prepend defaults to server options.
	opts = append([]NewServerOpt{ServerDefaults}, opts...)
	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

// RunServer handles the biolerplate of starting an http server and handling
// signals gracefully.
func RunServer(srv *http.Server) {
	idleConnsClosed := make(chan struct{})

	go func() {
		// Handle SIGINT and SIGTERM.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh
		fmt.Println("Received signal, stopping.", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	fmt.Printf("HTTP server listening on address: \"%s\"\n", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		fmt.Printf("HTTP server ListenAndServe: %v\n", err)
		os.Exit(1)
	}

	<-idleConnsClosed
}

// Env holds global dependencies that need to be initialized in main() and
// injected as dependencies into an application.
type Env struct {
	Reporter reporter.Reporter
	Logger   logger.Logger
	Context  context.Context
	Close    func() // Should be called in a defer in main().
}

// InitAll will initialize all the common dependencies such as metrics, reporting,
// tracing, and logging.
func InitAll() Env {
	traceCloser := InitTracer()
	metricsCloser := InitMetrics()

	l := InitLogger()
	logger.DefaultLogger = l

	r := InitReporter()

	ctx := reporter.WithReporter(context.Background(), r)
	ctx = logger.WithLogger(ctx, l)

	go func() {
		defer reporter.Monitor(ctx)
		metrics.Runtime()
	}()

	return Env{
		Logger:   l,
		Reporter: r,
		Context:  ctx,
		Close: func() {
			traceCloser()
			metricsCloser()
			reporter.Monitor(ctx)
		},
	}
}

// InitTracer configures a global datadog tracer.
//
// Env Vars:
// * DDTRACE_ADDR - The host:port of the local trace agent server.
// * EMPIRE_APPNAME - App name, used to construct the service name.
// * EMPIRE_PROCESS - Process name, used to construct the service name.
func InitTracer() func() {
	// create a Tracer configuration
	config := ddtrace.NewConfiguration()
	config.ServiceName = fmt.Sprintf("%s.%s", os.Getenv("EMPIRE_APPNAME"), os.Getenv("EMPIRE_PROCESS"))
	if addr := os.Getenv("DDTRACE_ADDR"); addr != "" {
		config.AgentHostname = addr
	}

	// Initialize a Tracer and ensure a graceful shutdown
	// using the `closer.Close()`
	tracer, closer, err := ddtrace.NewTracer(config)
	if err != nil {
		fmt.Println(err)
	}

	// set the Datadog tracer as a GlobalTracer
	opentracing.SetGlobalTracer(tracer)

	return func() {
		closer.Close()
	}
}

// InitMetrics configures pkg/metrics
//
// Env Vars:
// * STATSD_ADDR - The host:port of the statsd server.
func InitMetrics() func() {
	if addr := os.Getenv("STATSD_ADDR"); addr != "" {
		metrics.SetEmpireDefaultTags()
		metrics.Reporter, _ = metrics.NewDataDogMetricsReporter(addr)
	}

	return func() {
		metrics.Close()
	}
}

// InitLogger configures a leveled logger.
//
// Env Vars:
// * LOG_LEVEL - The log level
//
// If you want to replace the global default logger:
//	logger.DefaultLogger = InitLogger()
func InitLogger() logger.Logger {
	lvl := logger.ERROR
	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		lvl = logger.ParseLevel(ll)
	}

	return logger.New(log.New(os.Stdout, "", 0), lvl)
}

// InitReporter configures and returns a reporter.Reporter instance.
//
// Env Vars:
// * ROLLBAR_ACCESS_TOKEN - The Rollbar access token
// * ROLLBAR_ENVIRONMENT  - The Rollbar environment (staging, production)
// * ROLLBAR_ENDPOINT     - The Rollbar endpoint: https://api.rollbar.com/api/1/item/
func InitReporter() reporter.Reporter {
	rep := reporter.MultiReporter{}

	// Log Reporter, uses package level logger.
	rep = append(rep, reporter.NewLogReporter())

	// Rollbar reporter
	if os.Getenv(rollbar.EnvAccessToken) != "" && os.Getenv(rollbar.EnvEnvironment) != "" {
		rollbar.ConfigureFromEnvironment()
		rep = append(rep, rollbar.Reporter)
	} else {
		fmt.Println("Rollbar is not configured, skipping Rollbar reporter")
	}

	return rep
}
