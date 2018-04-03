package httpx

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/remind101/pkg/reporter"
)

func Error(ctx context.Context, err error, rw http.ResponseWriter, r *http.Request) {
	reporter.Report(ctx, err)
	EncodeError(err, rw)
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

func EncodeError(err error, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(ErrorStatusCode(err))

	errorResp := map[string]string{
		"error": err.Error(),
	}

	json.NewEncoder(rw).Encode(errorResp)
}

func ErrorStatusCode(err error) int {
	rootErr := errors.Cause(err)
	if e, ok := rootErr.(statusCoder); ok {
		return e.StatusCode()
	}
	if e, ok := rootErr.(temporaryError); ok && e.Temporary() {
		return http.StatusServiceUnavailable
	}

	if e, ok := rootErr.(timeoutError); ok && e.Timeout() {
		return http.StatusServiceUnavailable
	}

	return http.StatusInternalServerError
}
