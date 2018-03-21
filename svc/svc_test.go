package svc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/remind101/pkg/httpx"
	"github.com/remind101/pkg/httpx/middleware"
	"github.com/remind101/pkg/reporter/mock"
	"github.com/remind101/pkg/svc"
)

func TestStandardHandler(t *testing.T) {
	rep := mock.NewReporter()
	r := httpx.NewRouter()

	r.Handle("/panic", httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		panic("I panicked")
	}))

	r.Handle("/timeout", httpx.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}))

	h := svc.NewStandardHandler(svc.HandlerOpts{
		Router:         r,
		Reporter:       rep,
		ErrorHandler:   middleware.JSONReportingErrorHandler,
		HandlerTimeout: 500 * time.Millisecond,
	})

	s := httptest.NewServer(h)
	defer s.Close()

	// Test Panic
	req, _ := http.NewRequest("GET", s.URL+"/panic", nil)
	req.Header.Add("X-Request-ID", "abc")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := resp.StatusCode, 500; got != want {
		t.Errorf("got %d; expected %d", got, want)
	}

	if got, want := buf.String(), " request_id=abc error=\"I panicked\" line=22 file=svc_test.go\n"; got != want {
		t.Errorf("got %s; expected %s", got, want)
	}
}
