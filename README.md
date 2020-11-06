# [remind101/pkg](https://github.com/remind101/pkg) [![CircleCI](https://circleci.com/gh/remind101/pkg.svg?style=svg)](https://circleci.com/gh/remind101/pkg) [![GoDoc](https://godoc.org/github.com/remind101/pkg?status.svg)](https://godoc.org/github.com/remind101/pkg)

package pkg is a collection of Go packages that provide a layer of convenience over the stdlib.

## Packages

### [client](./client)

Helps build http clients with standard functionality such as error handling, tracing, timeouts,
request signatures, json encoding and decoding, etc.

### [counting](./counting)

Implements an a linear-time counting algorithm, also known as "linear counting".

### [httpmock](./httpmock)

A simple mock server implementation, useful for mocking external services in tests.

### [httpx](./httpx)

Defines the httpx.Handler interface, an httpx.Handler router, and a variety of middleware.

### [logger](./logger)

Defines a context aware structured leveled logger.

### [metrics](./metrics)

Defines an interface for metrics, with an implementation for Datadog.

### [reporter](./reporter)

Defines an interface for error reporting, with implementations for honeybadger, newrelic, and rollbar.

### [retry](./retry)

Provides the ability to retry a function call with exponential backoff and notification hooks.

### [stream](./stream)

Provides types that make it easier to perform streaming IO.

### [svc](./svc)

Provides tooling that integrates with all of the above packages to build
http servers with logging, distributed tracing, error reporting and metrics out of the box.

### [timex](./timex)

Use instead of time.Now to allow easy stubbing in tests.

### [profiling](./profiling)

Lightweight convenience stuff around
[pprof](https://golang.org/pkg/runtime/pprof/) and
[Google Cloud Profiler](https://cloud.google.com/profiler/docs) for visibility
into all our apps.

## Usage

For examples of usage, see the **[example](./example)** directory.
