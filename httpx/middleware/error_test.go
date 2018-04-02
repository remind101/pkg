package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/reporter"
)

type tmpError string

func (te tmpError) Error() string {
	return string(te)
}

func (te tmpError) Temporary() bool {
	return true
}

type statusCodeError struct {
	Err        error
	statusCode int
}

func (s statusCodeError) Error() string {
	return s.Err.Error()
}

func (s statusCodeError) StatusCode() int {
	return s.statusCode
}

func TestErrorMiddleware(t *testing.T) {
	tests := []struct {
		Error        error
		Body         string
		Code         int
		ErrorHandler ErrorHandlerFunc
	}{
		{
			Error: errors.New("boom"),
			Body:  "boom\n",
			Code:  500,
		},
		{
			Error: tmpError("service unavailable"),
			Body:  "service unavailable\n",
			Code:  503,
		},
		{
			Error: &net.DNSError{Err: "no such host", IsTimeout: true},
			Body:  "lookup : no such host\n",
			Code:  503,
		},
		{
			Error: statusCodeError{Err: errors.New("invalid request"), statusCode: 400},
			Body:  "invalid request\n",
			Code:  400,
		},
		{
			Error:        errors.New("boom"),
			Body:         "{\"error\":\"boom\"}\n",
			Code:         500,
			ErrorHandler: JSONReportingErrorHandler,
		},
	}

	for _, tt := range tests {
		h := &Error{
			handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return tt.Error
			}),
			ErrorHandler: tt.ErrorHandler,
		}
		req, _ := http.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		ctx := reporter.WithReporter(context.Background(), reporter.NewLogReporter())
		err := h.ServeHTTPContext(ctx, resp, req)
		if err != tt.Error {
			t.Fatal("Expected error to be returned")
		}

		if got, want := resp.Body.String(), tt.Body; got != want {
			t.Fatalf("Body => %#v; want %#v", got, want)
		}

		if got, want := resp.Code, tt.Code; got != want {
			t.Fatalf("Status => %v; want %v", got, want)
		}
	}
}

func TestErrorWithHandler(t *testing.T) {
	var called bool

	h := &Error{
		ErrorHandler: func(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
			called = true
		},
		handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return errors.New("boom")
		}),
	}

	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/path", nil)
	resp := httptest.NewRecorder()

	h.ServeHTTPContext(ctx, resp, req)

	if !called {
		t.Fatal("Expected the error handler to be called")
	}
}
