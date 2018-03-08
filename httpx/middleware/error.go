package middleware

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
)

type ErrorHandlerFunc func(context.Context, error, http.ResponseWriter, *http.Request)

type temporaryError interface {
	Temporary() bool // Is the error temporary?
}

type timeoutError interface {
	Timeout() bool // Is the error a timeout?
}

type statusCoder interface {
	StatusCode() int
}

// DefaultErrorHandler is an error handler that will respond with the error
// message and a 500 status.
var DefaultErrorHandler = func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	writeError(w, err)
}

// ReportingErrorHandler is an error handler that will report the error and respond
// with the error message and a 500 status.
var ReportingErrorHandler = func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	reporter.Report(ctx, err)
	writeError(w, err)
}

var JSONReportingErrorHandler = func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	reporter.Report(ctx, err)
	status := statusCodeForError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResp := map[string]string{
		"error": err.Error(),
	}

	json.NewEncoder(w).Encode(errorResp)
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), statusCodeForError(err))
}

func statusCodeForError(err error) int {
	if e, ok := err.(statusCoder); ok {
		return e.StatusCode()
	}
	if e, ok := err.(temporaryError); ok && e.Temporary() {
		return http.StatusServiceUnavailable
	}

	if e, ok := err.(timeoutError); ok && e.Timeout() {
		return http.StatusServiceUnavailable
	}

	return http.StatusInternalServerError
}

// Error is an httpx.Handler that will handle errors with an ErrorHandler.
type Error struct {
	// ErrorHandler is a function that will be called when a handler returns
	// an error.
	ErrorHandler ErrorHandlerFunc

	// Handler is the wrapped httpx.Handler that will be called.
	handler httpx.Handler
}

func NewError(h httpx.Handler) *Error {
	return &Error{
		handler: h,
	}
}

// HandleError returns a new Error middleware that uses f as the ErrorHandler.
func HandleError(h httpx.Handler, f ErrorHandlerFunc) *Error {
	e := NewError(h)
	e.ErrorHandler = f
	return e
}

// ServeHTTPContext implements the httpx.Handler interface.
func (h *Error) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	err := h.handler.ServeHTTPContext(ctx, w, r)

	if err != nil {
		f := h.ErrorHandler
		if f == nil {
			f = DefaultErrorHandler
		}

		f(ctx, err, w, r)
	}

	return nil
}
