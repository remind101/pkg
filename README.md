# [remind101/pkg](https://github.com/remind101/pkg) [![Build Status](https://travis-ci.org/remind101/pkg.svg?branch=master)](https://travis-ci.org/remind101/pkg) [![GoDoc](https://godoc.org/github.com/remind101/pkg?status.svg)](https://godoc.org/github.com/remind101/pkg)

package pkg is a collection of Go packages that provide a layer of convenience over the stdlib and primarily adds **[context.Context](https://godoc.org/golang.org/x/net/context)** awareness, making it easier to do things like distributed request tracing.

## Packages

* **[httpx](./httpx)**: Defines an httpx.Handler interface, an httpx.Handler router, and middleware.
* **[logger](./logger)**: Defines a context aware structured logger.
* **[reporter](./reporter)**: Defines a general abstraction for error reporting, with implementations for honeybadger.
* **[metrics](./metrics)**: Reports metrics to statsd (supports DataDog's metrics tagging)
* **[metricshttpx](./metrics/metricshttpx)**: Defines httpx-compatible middleware to report response times using `metrics`
* **[metricsmartini](./metrics/metricsmartini)**: Defines martini route-level middleware to report response times using `metrics`
* **[retry](./retry)**: Retry function calls with exponential backoff

## Usage

For examples of usage, see the **[example](./example)** directory.

## Why context.Context?

In dynamic languages like Ruby, it's common to use thread local variables to store request specific information, like a request id:

```ruby
Thread.current[:request_id] = env['HTTP_X_REQUEST_ID']
```

This isn't possible to do in Go. There are various implementations that try to overcome this by storing request specific information in a global variable, but this is problematic for a number of reasons.

1. Most of these implementations, like **[gorilla/context](https://github.com/gorilla/context)** tie the request specific information to an `*http.Request`, which is janky if your api is provided over something like a queue or rpc instead of http.
2. These implementations don't handle deadlines and cancellations in a generic way.

Using context.Context provides a generic way to allow request specific information to traverse api boundaries.
