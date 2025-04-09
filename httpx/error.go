package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/remind101/pkg/reporter"
)

// Error reports an error and encodes it to the http.ResponseWriter
func Error(ctx context.Context, err error, rw http.ResponseWriter, r *http.Request) {
	reporter.Report(ctx, err)
	EncodeError(err, rw)
}

// ErrorWithStatus reports an error with a specific status code
func ErrorWithStatus(ctx context.Context, err error, status int, rw http.ResponseWriter, r *http.Request) {
	reporter.Report(ctx, err)
	EncodeErrorWithStatus(err, status, rw)
}

type temporaryError interface {
	Temporary() bool // Is the error temporary?
}

type timeoutError interface {
	Timeout() bool // Is the error a timeout?
}

type statusCoder interface {
	StatusCode() int
}

// EncodeError encodes an error to the http.ResponseWriter
func EncodeError(err error, rw http.ResponseWriter) {
	EncodeErrorWithStatus(err, ErrorStatusCode(err), rw)
}

// EncodeErrorWithStatus encodes an error with a specific status code
func EncodeErrorWithStatus(err error, status int, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)

	errorResp := map[string]string{
		"error": err.Error(),
	}

	json.NewEncoder(rw).Encode(errorResp)
}

// ErrorStatusCode returns an appropriate HTTP status code based on the error type
func ErrorStatusCode(err error) int {
	var sc statusCoder
	if errors.As(err, &sc) {
		return sc.StatusCode()
	}

	var te temporaryError
	if errors.As(err, &te) && te.Temporary() {
		return http.StatusServiceUnavailable
	}

	var to timeoutError
	if errors.As(err, &to) && to.Timeout() {
		return http.StatusServiceUnavailable
	}

	return http.StatusInternalServerError
}

// NewError creates a new error with a message and optional key-value pairs
func NewError(msg string, keyvals ...interface{}) error {
	return &httpError{
		msg:     msg,
		keyvals: keyvals,
	}
}

// httpError is a structured error type that can include key-value pairs
type httpError struct {
	msg     string
	keyvals []interface{}
	status  int
	err     error
}

// Error implements the error interface
func (e *httpError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
	}
	return e.msg
}

// WithStatus sets the HTTP status code for the error
func (e *httpError) WithStatus(status int) *httpError {
	e.status = status
	return e
}

// WithError wraps another error
func (e *httpError) WithError(err error) *httpError {
	e.err = err
	return e
}

// StatusCode implements the statusCoder interface
func (e *httpError) StatusCode() int {
	if e.status != 0 {
		return e.status
	}
	return http.StatusInternalServerError
}

// Unwrap implements the errors.Wrapper interface for Go 1.13+ error unwrapping
func (e *httpError) Unwrap() error {
	return e.err
}
