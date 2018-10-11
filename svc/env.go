package svc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/remind101/pkg/logger"
	"github.com/remind101/pkg/metrics"
	"github.com/remind101/pkg/reporter"
	"github.com/remind101/pkg/reporter/rollbar"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

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
	var opts []tracer.StartOption
	// create a Tracer configuration
	opts = append(opts, tracer.WithServiceName(fmt.Sprintf(
		"%s.%s",
		os.Getenv("EMPIRE_APPNAME"),
		os.Getenv("EMPIRE_PROCESS"))))
	if addr := os.Getenv("DDTRACE_ADDR"); addr != "" {
		opts = append(opts, tracer.WithAgentAddr(addr))
	}

	// Initialize a Tracer and ensure a graceful shutdown
	// using the `closer.Close()`
	tracer := opentracer.New(opts...)

	// set the Datadog tracer as a GlobalTracer
	opentracing.SetGlobalTracer(tracer)

	return func() {
		// I can't find a way to flush in the new tracer API
	}
}

// InitMetrics configures pkg/metrics
//
// Env Vars:
// * STATSD_ADDR - The host:port of the statsd server.
func InitMetrics() (fn func()) {
	fn = func() {
		metrics.Close()
	}

	addr := os.Getenv("STATSD_ADDR")
	if addr == "" {
		return
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		return
	}

	if len(addrs) == 0 {
		return
	}

	metrics.SetEmpireDefaultTags()
	metrics.Reporter, _ = metrics.NewDataDogMetricsReporter(net.JoinHostPort(addrs[0], port))

	return
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
