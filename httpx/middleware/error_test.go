package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/remind101/pkg/httpx"
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
		Error error
		Body  string
		Code  int
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
			Error: statusCodeError{Err: errors.New("invalid request"), statusCode: 400},
			Body:  "invalid request\n",
			Code:  400,
		},
	}

	for _, tt := range tests {
		h := &Error{
			handler: httpx.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return tt.Error
			}),
		}
		req, _ := http.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		err := h.ServeHTTPContext(context.Background(), resp, req)
		if err != nil {
			t.Fatal("Expected no error to be returned because it was handled")
		}

		if got, want := resp.Body.String(), tt.Body; got != want {
			t.Fatalf("Body => %s; want %s", got, want)
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
