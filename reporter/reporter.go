// package reporter provides a context.Context aware abstraction for shuttling
// errors and panics to third partys.
package reporter

import (
	"fmt"
	"net/http"
	"strings"

	"context"

	"github.com/remind101/pkg/errctx"
)

// DefaultLevel is the default level a Report uses when reporting an error.
const DefaultLevel = "error"

// DefaultMax is the default maximum number of lines to show from the stack trace.
var DefaultMax = 1024

// Reporter represents an error handler.
type Reporter interface {
	// Report reports the error to an external system. The provided error
	// could be an Error instance, which will contain additional information
	// about the error, including a stack trace and any contextual
	// information. Implementers should type assert the error to an *Error
	// if they want to report the stack trace.
	ReportWithLevel(context.Context, string, error) error
}

// ReporterFunc is a function signature that conforms to the Reporter interface.
type ReporterFunc func(context.Context, string, error) error

// Report implements the Reporter interface.
func (f ReporterFunc) ReportWithLevel(ctx context.Context, level string, err error) error {
	return f(ctx, level, err)
}

// FromContext extracts a Reporter from a context.Context.
func FromContext(ctx context.Context) (Reporter, bool) {
	h, ok := ctx.Value(reporterKey).(Reporter)
	return h, ok
}

// WithReporter inserts a Reporter into the context.Context.
func WithReporter(ctx context.Context, r Reporter) context.Context {
	return context.WithValue(ctx, reporterKey, r)
}

// AddContext adds contextual information to the Request object.
func AddContext(ctx context.Context, key string, value interface{}) context.Context {
	return errctx.WithInfo(ctx, key, value)
}

// AddRequest adds information from an http.Request to the Request object.
func AddRequest(ctx context.Context, req *http.Request) context.Context {
	return errctx.WithRequest(ctx, req)
}

// MultiError is an error implementation that wraps multiple errors.
type MultiError struct {
	Errors []error
}

// Error implements the error interface. It simply joins all of the individual
// error messages with a comma.
func (e *MultiError) Error() string {
	var m []string

	for _, err := range e.Errors {
		m = append(m, err.Error())
	}

	return strings.Join(m, ", ")
}

// ReportWithLevel wraps the err as an Error and reports it the the Reporter embedded
// within the context.Context.
func ReportWithLevel(ctx context.Context, level string, err error) error {
	e := errctx.New(ctx, err, 1)
	return reportWithLevel(ctx, level, e)
}

// Report wraps the err as an Error and reports it the the Reporter embedded
// within the context.Context.
func Report(ctx context.Context, err error) error {
	e := errctx.New(ctx, err, 1)
	return reportWithLevel(ctx, DefaultLevel, e)
}

func reportWithLevel(ctx context.Context, level string, err error) error {
	if r, ok := FromContext(ctx); ok {
		return r.ReportWithLevel(ctx, level, err)
	}

	panic("No reporter in provided context.")
}

// Monitors and reports panics. Useful in goroutines.
// Example:
//   ctx := reporter.WithReporter(context.Background(), hb2.NewReporter(hb2.Config{}))
//   ...
//   go func(ctx context.Context) {
//     defer reporter.Monitor(ctx)
//     ...
//     panic("oh noes") // will report, then crash.
//   }(ctx)
func Monitor(ctx context.Context) {
	if v := recover(); v != nil {
		var err error
		if e, ok := v.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", v)
		}
		Report(ctx, err)
		panic(err)
	}
}

// key used to store context values from within this package.
type key int

const (
	reporterKey key = iota
)
