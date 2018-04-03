package httpx_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestError(t *testing.T) {
	tests := []struct {
		Error error
		Body  string
		Code  int
	}{
		{
			Error: errors.New("boom"),
			Body:  `{"error":"boom"}` + "\n",
			Code:  500,
		},
		{
			Error: tmpError("service unavailable"),
			Body:  `{"error":"service unavailable"}` + "\n",
			Code:  503,
		},
		{
			Error: &net.DNSError{Err: "no such host", IsTimeout: true},
			Body:  `{"error":"lookup : no such host"}` + "\n",
			Code:  503,
		},
		{
			Error: statusCodeError{Err: errors.New("invalid request"), statusCode: 400},
			Body:  `{"error":"invalid request"}` + "\n",
			Code:  400,
		},
	}

	for _, tt := range tests {
		r, _ := http.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		ctx := reporter.WithReporter(context.Background(), reporter.NewLogReporter())
		httpx.Error(ctx, tt.Error, rw, r)

		if got, want := rw.Body.String(), tt.Body; got != want {
			t.Fatalf("Body => %#v; want %#v", got, want)
		}

		if got, want := rw.Code, tt.Code; got != want {
			t.Fatalf("Status => %v; want %v", got, want)
		}
	}
}
