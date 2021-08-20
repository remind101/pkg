// package errors provides error handling primitives in a request context.
//
// Adding request information
//
//     var ctx context.Context
//     var req *http.Request
//     ctx = errors.WithRequest(ctx, req)
//
// Adding contextual information
//
//     ctx = errors.WithInfo(ctx, "X-Request-ID", "123")
//
// Creating an error with context
//
//     e := errors.New(ctx, err, 0)
//     e.Err                                    // err
//     e.Request()                              // *http.Request
//     e.ContextData()["X-Request-ID"].(string) // "123"
//     e.StackTrace()                           // errors.StackTrace
package errors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// MaxFrames is the default maximum number of lines to show from the stack trace.
var MaxFrames = 1024

// WithInfo adds contextual information to the info object in the context.
func WithInfo(ctx context.Context, key string, value interface{}) context.Context {
	ctx = withInfo(ctx)
	i, _ := infoFromContext(ctx)
	i.data[key] = value
	return ctx
}

// WithRequest adds information from an http.Request to the info object in the context.
func WithRequest(ctx context.Context, req *http.Request) context.Context {
	ctx = withInfo(ctx)
	i, _ := infoFromContext(ctx)
	i.request = safeCloneRequest(req)
	return ctx
}

// Recover wraps the return value of recover() to capture a panic stack correctly.
func Recover(ctx context.Context, v interface{}) (e error) {
	switch err := v.(type) {
	case nil:
		e = nil
	case *Error:
		e = err
	case error:
		e = New(ctx, err, 0)
	default:
		e = New(ctx, fmt.Errorf("%v", err), 0)
	}

	return e
}

// Error wraps an error with additional information, like a stack trace,
// contextual information, and an http request if provided.
type Error struct {
	// The error that was generated.
	Err error

	// Any freeform contextual information about that error.
	info map[string]interface{}

	// If provided, an http request that generated the error.
	request *http.Request

	// This is private so that it can be exposed via StackTrace(),
	// which implements the stackTracker interface.
	stackTrace errors.StackTrace
}

// New returns a new Error instance. If err is already an Error instance,
// it will be returned, otherwise err will be wrapped with Error.
func New(ctx context.Context, err error, skip int) *Error {
	if e, ok := err.(*Error); ok {
		return e
	}
	return new(err, skip+1).WithContext(ctx)
}

// new wraps err as an Error and generates a stack trace pointing at the
// caller of this function.
func new(err error, skip int) *Error {
	return &Error{
		Err:        err,
		stackTrace: stacktrace(err, skip+1),
		info:       map[string]interface{}{},
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Err.Error()
}

// Cause implements the causer interface.
func (e *Error) Cause() error {
	return errors.Cause(e.Err)
}

// StackTrace implements the stackTracer interface.
func (e *Error) StackTrace() errors.StackTrace {
	return e.stackTrace
}

// Request returns the request object associated with this error.
func (e *Error) Request() *http.Request {
	return e.request
}

// ContextData() returns contextual information associated with this error.
func (e *Error) ContextData() map[string]interface{} {
	return e.info
}

// WithContext returns a new Error with contextual information added.
func (e *Error) WithContext(ctx context.Context) *Error {
	if i, ok := infoFromContext(ctx); ok {
		e.info = i.data
		e.request = i.request
	}
	return e
}

type causer interface {
	Cause() error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// It generates a brand new stack trace given an error and
// the number of frames that should be skipped,
// from innermost to outermost frames.
func genStacktrace(err error, skip int) errors.StackTrace {
	var stack errors.StackTrace
	errWithStack := errors.WithStack(err)
	stack = errWithStack.(stackTracer).StackTrace()
	skip++

	// if it is recovering from a panic() call,
	// reset the stack trace at that point
	for index, frame := range stack {
		file := fmt.Sprintf("%s", frame)
		if file == "panic.go" {
			skip = index + 1
			break
		}
	}
	if skip >= len(stack) {
		panic("attempt to skip past more frames than are present in the stack")
	}

	return stack[skip:]
}

// There are two interfaces that drive this implementation:
//
//   * causer
//     - it unwraps an error instance in a chain of errors created with errors.Wrap
//     - therefore, the last one in the chain is the root cause (inner-most)
//
//   * stackTracer
//     - not all errors in the aforementioned chain may have a stack trace,
//
// It returns the innermost stack trace in a chain of errors because it is
// the closest to the root cause.
//
func getStacktrace(err error) errors.StackTrace {
	var stack errors.StackTrace
	for err != nil {
		errWithStack, stackOK := err.(stackTracer)
		if stackOK && errWithStack.StackTrace() != nil {
			stack = errWithStack.StackTrace()
		}
		if errWithCause, causerOK := err.(causer); causerOK {
			err = errWithCause.Cause()
		} else {
			// end of chain
			break
		}
	}
	return stack
}

func stacktrace(err error, skip int) errors.StackTrace {
	stack := getStacktrace(err)
	if stack == nil {
		stack = genStacktrace(err, skip+1)
	}
	if len(stack) > MaxFrames {
		stack = stack[:MaxFrames]
	}
	return stack
}

// Will recover from a panic, and throw the error away.
//
// Useful for when you want to test that your panic handling is working
// correctly. If your test actually throws a panic, it just crashes right
// there, so you can use this function to ignore the panic.
//
// Ex:
// go func() {
//   defer errors.IgnorePanic()
//   defer errors.PushPanicToChannel(errChan)
//
//   DoSomethingDangerous()
// }()
func IgnorePanic() {
	recover()
}

// If there's a panic, this will push the error to the given error channel.
//
// Useful when you want to do many things async, and want any error put in
// a channel. A common mistake is to only handle returned errors, and forget
// to handle panics as well, leading to errors that look like timeout errors
// but are actually panics.
//
// Ex:
// go func() {
//   defer errors.PushPanicToChannel(errChan)
//
//   result, err := DoSomethingDangerous()
//   if err != nil {
//     errChan <- err
//   } else {
//     resultChan <- result
//   }
// }()
func PushPanicToChannel(ctx context.Context, errChan chan error) {
	if err := Recover(ctx, recover()); err != nil {
		errChan <- err
		panic(err)
	}
}
