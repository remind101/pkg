# [remind101/pkg](https://github.com/remind101/pkg)

package pkg is a collection of Go packages that provide a layer of convenience over the stdlib and primarily adds **[context.Context](https://godoc.org/golang.org/x/net/context)** awareness, making it easier to do things like distributed request tracing.

## Packages

* **[httpx](./httpx)**: Defines an httpx.Handler interface, an httpx.Handler router, and middleware.
* **[logger](./logger)**: Defines a context aware structured logger.
* **[reporter](./reporter)**: Defines a general abstraction for error reporting, with implementations for honeybadger.
* **[metrics](./metrics)**: TODO
