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
//		})
//
// 	svc.RunServer(h, "80", 5*time.Second)
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
	"github.com/pkg/errors"
	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/metrics"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/reporter/rollbar"
)

type HandlerOpts struct {
	Router    *httpx.Router
	Reporter  reporter.Reporter
	BasicAuth string
}

// NewStandardHandler
func NewStandardHandler(opts HandlerOpts) http.Handler {
	var h httpx.Handler

	// Recover from panics.
	h = middleware.Recover(opts.Router, opts.Reporter)

	// Add request tracing
	h = middleware.OpentracingTracing(h, opts.Router)

	// Add the request id to the context.
	h = middleware.ExtractRequestID(h)

	// Add basic auth
	if opts.BasicAuth != "" {
		user := strings.Split(opts.BasicAuth, ":")[0]
		pass := strings.Split(opts.BasicAuth, ":")[1]
		h = middleware.BasicAuth(h, user, pass, "")
	}

	// Wrap the route in middleware to add a context.Context.
	return middleware.BackgroundContext(h)
}

// RunServer handles the biolerplate of starting an http server and handling
// signals gracefully.
func RunServer(h http.Handler, port string, writeTimeout time.Duration) {
	errCh := make(chan error)

	// Handle SIGINT and SIGTERM.
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Listening on port %s\n", port)

	// Add timeouts to the server
	srv := &http.Server{
		WriteTimeout: writeTimeout * time.Second,
		Addr:         ":" + port,
		Handler:      h,
	}

	go func() {
		defer reporter.Monitor(context.Background())
		err := srv.ListenAndServe()
		if err != nil {
			errCh <- errors.Wrapf(err, "unable to start server")
		}
	}()

	select {
	case sig := <-sigCh:
		fmt.Println("Received signal, stopping.", "signal", sig)
	// Cleanup
	case err := <-errCh:
		fmt.Println(err)
		os.Exit(1)
	}
}

// Env holds global dependencies that need to be initialized in main() and
// injected as dependencies into an application.
type Env struct {
	Reporter reporter.Reporter
	Logger   logger.Logger
	Close    func() // Should be called in a defer in main().
}

// InitAll will initialize all the common dependencies such as metrics, reporting,
// tracing, and logging.
func InitAll() Env {
	traceCloser := InitTracer()
	metricsCloser := InitMetrics()

	return Env{
		Logger:   InitLogger(),
		Reporter: InitReporter(),
		Close: func() {
			traceCloser()
			metricsCloser()
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

	return logger.New(log.New(os.Stdout, "", log.Lshortfile), lvl)
}

// InitReporter configures and returns a reporter.Reporter instance.
//
// Env Vars:
// * ROLLBAR_ACCESS_TOKEN - The Rollbar access token
// * ROLLBAR_ENVIRONMENT  - The Rollbar environment (staging, production)
func InitReporter() reporter.Reporter {
	rbToken := os.Getenv("ROLLBAR_ACCESS_TOKEN")
	rbEnv := os.Getenv("ROLLBAR_ENVIRONMENT")

	rep := reporter.MultiReporter{}

	// Log Reporter
	rep = append(rep, reporter.NewLogReporter())

	// Rollbar reporter
	if rbToken != "" && rbEnv != "" {
		rollbar.ConfigureReporter(rbToken, rbEnv)
		rep = append(rep, rollbar.Reporter)
	} else {
		fmt.Println("Rollbar is not configured, skipping Rollbar reporter")
	}

	return rep
}
