// package reporter provides a context.Context aware abstraction for shuttling
// errors and panics to third partys.
package reporter

import (
	"net/http"
	"runtime"
	"strings"

	"golang.org/x/net/context"
)

// DefaultMax is the default maximum number of lines to show from the backtrace.
var DefaultMax = 1024

// Reporter represents an error handler.
type Reporter interface {
	// Report reports the error to an external system.
	Report(context.Context, error) error
}

// ReporterFunc is a function signature that conforms to the Reporter interface.
type ReporterFunc func(context.Context, error) error

// Report implements the Reporter interface.
func (f ReporterFunc) Report(ctx context.Context, err error) error {
	return f(ctx, err)
}

// FromContext extracts a Reporter from a context.Context.
func FromContext(ctx context.Context) (Reporter, bool) {
	h, ok := ctx.Value(reporterKey).(Reporter)
	return h, ok
}

// WithReporter inserts a Reporter into the context.Context.
func WithReporter(ctx context.Context, r Reporter) context.Context {
	return context.WithValue(withInfo(ctx), reporterKey, r)
}

// AddContext adds contextual information to the Request object.
func AddContext(ctx context.Context, key string, value interface{}) {
	i := infoFromContext(ctx)
	i.context[key] = value
}

// AddRequest adds information from an http.Request to the Request object.
func AddRequest(ctx context.Context, req *http.Request) {
	i := infoFromContext(ctx)
	// TODO clone the request?
	i.request = req
}

// Report generates a new ErrorReport and uses the embedded reporter to report
// it.
func Report(ctx context.Context, err error) error {
	e := NewErrorWithContext(ctx, err, 1)

	if r, ok := FromContext(ctx); ok {
		return r.Report(ctx, e)
	}

	return nil
}

// A line from the backtrace.
type BacktraceLine struct {
	PC   uintptr
	File string
	Line int
}

// Error represents information about an error.
type Error struct {
	// The error that was generated.
	Err error

	// The backtrace.
	Backtrace []*BacktraceLine

	// Any freeform contextual information about that error.
	Context map[string]interface{}

	// If provided, an http request that generated the error.
	Request *http.Request
}

// Make error implement the error interface.
func (e *Error) Error() string {
	return e.Err.Error()
}

// NewError returns a new Error instance.
func NewError(err error, skip int) *Error {
	return &Error{
		Err:       err,
		Backtrace: backtrace(skip+1, DefaultMax),
	}
}

// NewErrorWithContext returns a new Error with contextual information added.
func NewErrorWithContext(ctx context.Context, err error, skip int) *Error {
	e := NewError(err, skip+1)
	i := infoFromContext(ctx)
	e.Context = i.context
	e.Request = i.request
	return e
}

// MutliError is an error implementation that wraps multiple errors.
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

// backtrace generates a backtrace and returns a slice of BacktraceLine.
func backtrace(skip, max int) []*BacktraceLine {
	var lines []*BacktraceLine

	for i := skip + 1; i < max; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		lines = append(lines, &BacktraceLine{
			PC:   pc,
			File: file,
			Line: line,
		})
	}

	return lines
}

// info is used internally to store contextual information. Any empty info
// gets inserted into the context.Context when the Reporter is inserted, which
// allows downstream consumers to add additional information to this object.
type info struct {
	context map[string]interface{}
	request *http.Request
}

func newInfo() *info {
	return &info{context: make(map[string]interface{})}
}

func withInfo(ctx context.Context) context.Context {
	return context.WithValue(ctx, infoKey, newInfo())
}

func infoFromContext(ctx context.Context) *info {
	i, ok := ctx.Value(infoKey).(*info)
	if !ok {
		return newInfo()
	}
	return i
}

// key used to store context values from within this package.
type key int

const (
	reporterKey key = iota
	infoKey
)
